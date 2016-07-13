package dtb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()
// ConvertDataToBytes convert any data to byte array
func ConvertDataToBytes(data interface{}, endian binary.ByteOrder) ([]byte, error) {
	dataValue := reflect.ValueOf(data)
	return convertValueToBytes(dataValue, endian)
}
func convertValueToBytes(value reflect.Value, endian binary.ByteOrder) ([]byte, error) {
	var err error
	result := bytes.Buffer{}
	valType := value.Type()
	switch valType.Kind() {
	case reflect.Array:
		arrayLength := valType.Len()
		for i := 0; i < arrayLength; i++ {
			var val []byte
			val, err = convertValueToBytes(value.Index(i), endian)
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
		return nil, fmt.Errorf("Unsupported type `%s` to convert to bytes\n", valType.Kind().String())
	case reflect.Ptr:
		return convertValueToBytes(value.Elem(), endian)
	default:
		if value.CanInterface() {
			binary.Write(&result, endian, value.Interface())
		} else {
			return nil, fmt.Errorf("Can't get value from %s", valType.Name())
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
		sFuncs := fieldType.Tag.Get("bytes_fn")
		if sFuncs != "" {
			funcs := strings.Split(sFuncs, ",")
			if len(funcs) < 2 {
				return nil, fmt.Errorf("You should specify two function names separated by comma in `bytes_fn` in field %s", fieldType.Name)
			}
			methodName := funcs[0]
			methodType, ok := sType.MethodByName(methodName)
			if !ok {
				return nil, fmt.Errorf("Structure %s doesn't have method `%s` to encode `%s` to bytes (Check, maybe this method have pointer receiver)", sType.Name(), funcs[0], fieldType.Name)
			}

			data, err := encodeValueViaFunc(sType.Name(), fieldValue, value.MethodByName(methodName), methodType)
			if err != nil {
				return nil, err
			}
			binary.Write(&result, endian, data)
			continue
		}
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
				fieldByteValue, err = convertValueToBytes(fieldValue, endian)
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

func encodeValueViaFunc(structName string, value reflect.Value, method reflect.Value, methodType reflect.Method) ([]byte, error) {
	methodName := methodType.Name
	if methodType.Type.NumIn() != 1 {
		return nil, fmt.Errorf("Method %s.%s should not receive any arguments", structName, methodName)
	}
	if methodType.Type.NumOut() != 2 {
		return nil, fmt.Errorf("Method %s.%s should return 2 values ([]byte and error)", structName, methodName)
	}
	if methodType.Type.Out(0) != reflect.TypeOf([]byte{}) {
		return nil, fmt.Errorf("Method's %s.%s first return value should be []byte{}", structName, methodName)
	}
	if !methodType.Type.Out(1).Implements(errorInterface) {
		return nil, fmt.Errorf("Method's %s.%s second return value should be error(current:%v)", structName, methodName, methodType.Type.Out(1))
	}
	values := method.Call([]reflect.Value{})
	if err, _ := values[1].Interface().(error); err != nil {
		return nil, err
	}
	result, _ := values[0].Interface().([]byte)
	return result, nil
}