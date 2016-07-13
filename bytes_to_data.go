package dtb

import (
	"encoding/binary"
	"reflect"
	"errors"
	"math"
	"strconv"
	"fmt"
)
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