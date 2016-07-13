package structsToBytes

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
)

// ConvertDataToBytes convert any data to byte array
func ConvertDataToBytes(data interface{}, endian binary.ByteOrder) ([]byte, error) {
	dataType := reflect.TypeOf(data)
	dataValue := reflect.ValueOf(data)
	return convertValueToBytes(dataValue, dataType, endian)
}
func convertValueToBytes(value reflect.Value, Type reflect.Type, endian binary.ByteOrder) ([]byte, error) {
	var err error
	result := bytes.Buffer{}
	switch Type.Kind() {
	case reflect.Array:
		arrayElementsType := Type.Elem()
		arrayLength := Type.Len()
		for i := 0; i < arrayLength; i++ {
			var val []byte
			val, err = convertValueToBytes(value.Index(i), arrayElementsType, endian)
			if err != nil {
				return nil, err
			}
			binary.Write(&result, endian, val)
		}
	case reflect.Struct:
		fieldsCount := Type.NumField()
		for i := 0; i < fieldsCount; i++ {
			fieldType := Type.Field(i)
			fieldValue := value.Field(i)
			ignoreField := fieldType.Tag.Get("bytes_ignore")
			if ignoreField != "" {
				needIgnoreField, err := strconv.ParseBool(ignoreField)
				if err == nil && needIgnoreField {
					continue
				}
			}
			var fieldByteValue interface{}
			if fieldValue.CanInterface() {
				if fieldType.Type.Kind() == reflect.String {
					strValue := fieldValue.String()
					strLength := fieldType.Tag.Get("bytes_length")
					length, err := strconv.ParseInt(strLength, 10, 32)
					if err != nil {
						log.Panicf("You should specify strings length (tag `bytes_length`) for field `%s`\n", fieldType.Name)
					}
					val := make([]byte, length)
					copy(val, strValue)
					fieldByteValue = val
				} else {
					fieldByteValue, err = convertValueToBytes(fieldValue, fieldType.Type, endian)
					if err != nil {
						return nil, err
					}
				}
			} else {
				fieldByteValue = make([]byte, typeSize(fieldType.Type))
			}
			//fmt.Printf("%x - %s\n", result.Len(), field_type.Name)
			binary.Write(&result, endian, fieldByteValue)
		}
	case reflect.String, reflect.Slice:
		log.Panicf("Unsupported type `%s` to convert to bytes\n", Type.Kind().String())
	default:
		if value.CanInterface() {
			binary.Write(&result, endian, value.Interface())
		} else {
			return nil, fmt.Errorf("Can't get value from %s", Type.Name())
		}
	}
	return result.Bytes(), nil
}

// ConvertBytesToData write byte array to data
func ConvertBytesToData(bytes []byte, endian binary.ByteOrder, data interface{}) error {
	dataType := reflect.TypeOf(data)
	if dataType.Kind() != reflect.Ptr {
		return errors.New("Data should be pointer")
	}
	dataType = dataType.Elem()
	dataValue := reflect.ValueOf(data).Elem()
	_, err := updateValueByTypeFromBytes(dataValue, dataType, bytes, endian)
	return err
}

func updateValueByTypeFromBytes(value reflect.Value, Type reflect.Type, bytes []byte, endian binary.ByteOrder) (offset int, err error) {
	switch Type.Kind() {
	case reflect.Int8:
		value.SetInt(int64(int8(bytes[0])))
		offset = 1
	case reflect.Int16:
		val := endian.Uint16(bytes[:2])
		value.SetInt(int64(int16(val)))
		offset = 2
	case reflect.Int32:
		val := endian.Uint32(bytes[:4])
		value.SetInt(int64(int32(val)))
		offset = 4
	case reflect.Int64:
		val := endian.Uint64(bytes[:8])
		value.SetInt(int64(val))
		offset = 8
	case reflect.Uint8:
		value.SetUint(uint64(bytes[offset]))
		offset = 1
	case reflect.Uint16:
		val := endian.Uint16(bytes[:2])
		value.SetUint(uint64(int16(val)))
		offset = 2
	case reflect.Uint32:
		val := endian.Uint32(bytes[:4])
		value.SetUint(uint64(int16(val)))
		offset = 4
	case reflect.Uint64:
		value.SetUint(endian.Uint64(bytes[:8]))
		offset = 8
	case reflect.Float32:
		val := endian.Uint32(bytes[:4])
		float := math.Float32frombits(val)
		value.SetFloat(float64(float))
		offset = 4
	case reflect.Float64:
		val := endian.Uint64(bytes[:8])
		float := math.Float64frombits(val)
		value.SetFloat(float)
		offset = 8
	case reflect.Struct:
		fieldsCount := Type.NumField()
		for i := 0; i < fieldsCount; i++ {
			fieldType := Type.Field(i)
			fieldValue := value.Field(i)
			ignoreField := fieldType.Tag.Get("bytes_ignore")
			if !fieldValue.CanInterface() {
				offset += typeSize(fieldType.Type)
				continue
			}
			if ignoreField != "" {
				needIgnoreField, err := strconv.ParseBool(ignoreField)
				if err == nil && needIgnoreField {
					continue
				}
			}
			if fieldType.Type.Kind() == reflect.String {
				strLength := fieldType.Tag.Get("bytes_length")
				length, err := strconv.ParseInt(strLength, 10, 32)
				if err != nil {
					return 0, fmt.Errorf("You should specify strings length (tag `bytes_length`) for field `%s`", fieldType.Name)
				}
				fieldValue.SetString(bytesToStr(bytes[offset : offset+int(length)]))
				offset += int(length)

			} else {
				newOffset, err := updateValueByTypeFromBytes(fieldValue, fieldType.Type, bytes[offset:], endian)
				if err != nil {
					return 0, err
				}
				offset += newOffset
			}
		}
	case reflect.Array, reflect.Slice:
		arrayItemsType := Type.Elem()
		arrayLength := value.Len()
		for i := 0; i < arrayLength; i++ {
			newOffset, err := updateValueByTypeFromBytes(value.Index(i), arrayItemsType, bytes[offset:], endian)
			if err != nil {
				return 0, err
			}
			offset += newOffset
		}
	case reflect.Interface:
		interfaceValue := value.Elem()
		interfaceType := interfaceValue.Type()

		newOffset, err := updateValueByTypeFromBytes(interfaceValue, interfaceType, bytes[offset:], endian)
		if err != nil {
			return 0, err
		}
		offset += newOffset
	default:
		return 0, fmt.Errorf("Type %v is not supported yet.\n", Type.Kind())
	}
	return offset, nil
}
func typeSize(Type reflect.Type) int {
	switch Type.Kind() {
	case reflect.Array:
		return typeSize(Type.Elem()) * Type.Len()
	default:
		return Type.Bits() / 8
	}
}
func bytesToStr(bytes []byte) string {
	for key, value := range bytes {
		if value == '\u0000' {
			return string(bytes[:key])
		}
	}
	return string(bytes[:])
}
