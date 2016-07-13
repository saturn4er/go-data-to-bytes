package dtb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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

