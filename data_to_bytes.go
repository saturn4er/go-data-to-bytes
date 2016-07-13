package dtb

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
		structBytes, err := convertStructToBytes(value, endian)
		if err != nil {
			return nil, err
		}

		binary.Write(&result, endian, structBytes)
	case reflect.String, reflect.Slice:
		return nil, fmt.Errorf("Unsupported type `%s` to convert to bytes\n", Type.Kind().String())
	default:
		if value.CanInterface() {
			binary.Write(&result, endian, value.Interface())
		} else {
			return nil, fmt.Errorf("Can't get value from %s", Type.Name())
		}
	}
	return result.Bytes(), nil
}

func convertStructToBytes(value reflect.Value, endian binary.ByteOrder) ([]byte, error) {
	var err error
	sType := value.Type()
	fieldsCount := sType.NumField()
	result := bytes.Buffer{}
	for i := 0; i < fieldsCount; i++ {
		fieldValue := value.Field(i)
		fieldType := sType.Field(i)
		// Check if we should ignore field
		sIgnoreField := fieldType.Tag.Get("bytes_ignore")
		if sIgnoreField != "" {
			var ignoreField bool
			ignoreField, err = strconv.ParseBool(sIgnoreField)
			if err == nil && ignoreField {
				continue
			}
		}
		var fieldByteValue interface{}
		if fieldValue.CanInterface() {
			if fieldType.Type.Kind() == reflect.String {
				strValue := fieldValue.String()
				strLength := fieldType.Tag.Get("bytes_length")
				var length int64
				length, err = strconv.ParseInt(strLength, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("You should specify valid `bytes_length` tag for %s field of type string", fieldType.Name)
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

		binary.Write(&result, endian, fieldByteValue)
	}
	return result.Bytes(), nil
}