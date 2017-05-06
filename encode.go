package d2b

import (
	"bytes"
	"encoding/binary"
	"reflect"

	"github.com/pkg/errors"
)

// Encode converts interface type to bytes array
func Encode(data interface{}, endian binary.ByteOrder) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := valueToBytes(reflect.ValueOf(data), buffer, endian)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// getTypeBytesLength returns reflect.Type's bytes representation
func valueToBytes(v reflect.Value, buffer *bytes.Buffer, endian binary.ByteOrder) error {
	kind := v.Kind()
	t := v.Type()
	switch kind {
	case reflect.Ptr:
		if v.IsNil() {
			typeLen, err := getTypeBytesLength(v.Type().Elem())
			if err != nil {
				return err
			}
			buffer.Write(make([]byte, typeLen))
			return nil
		}
		return valueToBytes(v.Elem(), buffer, endian)
	case reflect.Struct:
		tags, err := getStructTags(t)
		if err != nil {
			return errors.Wrapf(err, "parsing %v struct tags error", t.Name())
		}
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			err := structFieldValueToBytes(v.Field(i), tags[i], buffer, endian)
			if err != nil {
				return errors.Wrapf(err, "can't encode %v.%v field to bytes", t.Name(), ft.Name)
			}
		}
		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool:
		return binary.Write(buffer, endian, v.Interface())
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := valueToBytes(v.Index(i), buffer, endian)
			if err != nil {
				return errors.Wrap(err, "can't convert array element to bytes")
			}
		}
		return nil
	}
	return errors.New("unsupported type: " + kind.String())
}
func structFieldValueToBytes(v reflect.Value, ft *structFieldTag, buffer *bytes.Buffer, endian binary.ByteOrder) error {
	if ft.Skip {
		return nil
	}
	k := v.Kind()
	switch k {
	case reflect.Ptr:
		if v.IsNil() {
			typeLen, err := getStructFieldTypeBytesLength(v.Type().Elem(), ft)
			if err != nil {
				return err
			}
			buffer.Write(make([]byte, typeLen))
			return nil
		}
		return structFieldValueToBytes(v.Elem(), ft, buffer, endian)
	case reflect.String:
		if ft.Length == 0 {
			return errors.New("need to specify length")
		}
		val := v.String()
		b := make([]byte, ft.Length)
		copy(b, val)
		buffer.Write(b)
	case reflect.Slice:
		if ft.Length == 0 {
			return errors.New("need to specify length")
		}

		var l = v.Len()
		var handleLength = ft.Length
		if l < handleLength {
			handleLength = l
		}
		for i := 0; i < handleLength; i++ {
			err := valueToBytes(v.Index(i), buffer, endian)
			if err != nil {
				return errors.Wrap(err, "can't convert slice element to bytes")
			}
		}
		if handleLength < ft.Length {
			typeLen, err := getTypeBytesLength(v.Type().Elem())
			if err != nil {
				return errors.Wrap(err, "can't calculate slice element type length")
			}
			placeholder := make([]byte, typeLen*(ft.Length-handleLength))
			buffer.Write(placeholder)
		}

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := valueToBytes(v.Index(i), buffer, endian)
			if err != nil {
				return errors.Wrap(err, "can't convert array element to bytes")
			}
		}
	default:
		return valueToBytes(v, buffer, endian)
	}
	return nil
}

// getTypeBytesLength returns reflect.Type's length in bytes
func getTypeBytesLength(t reflect.Type) (int, error) {
	kind := t.Kind()
	switch kind {
	case reflect.Ptr:
		return getTypeBytesLength(t.Elem())
	case reflect.Struct:
		var result int
		tags, err := getStructTags(t)
		if err != nil {
			return 0, errors.Wrapf(err, "parsing %v struct tags error", t.Name())
		}
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			fl, err := getStructFieldTypeBytesLength(ft.Type, tags[i])
			if err != nil {
				return 0, errors.Wrapf(err, "detecting %v.%v field length error", t.Name(), ft.Name)
			}

			result += fl
		}
		return result, nil
	case reflect.Int8, reflect.Uint8:
		return 1, nil
	case reflect.Int16, reflect.Uint16:
		return 2, nil
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 4, nil
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return 8, nil
	case reflect.Array:
		elLen, err := getTypeBytesLength(t.Elem())
		if err != nil {
			return 0, errors.Wrap(err, "detecting array element type length error")
		}
		return t.Len() * elLen, nil
	}
	return 0, errors.New("unsupported type: " + kind.String())
}

// getTypeBytesLength returns reflect.Type's length in bytes, relying on struct tag
func getStructFieldTypeBytesLength(r reflect.Type, tagInfo *structFieldTag) (int, error) {
	if tagInfo.Skip {
		return 0, nil
	}
	switch r.Kind() {
	case reflect.Ptr:
		return getStructFieldTypeBytesLength(r.Elem(), tagInfo)
	case reflect.Slice:
		if tagInfo.Length == 0 {
			return 0, errors.New("need to specify length")
		}
		elemLength, err := getTypeBytesLength(r.Elem())
		if err != nil {
			return 0, errors.Wrap(err, "can't detect slice element length")
		}
		return tagInfo.Length * elemLength, nil
	case reflect.Array:
		elemLength, err := getTypeBytesLength(r.Elem())
		if err != nil {
			return 0, errors.Wrap(err, "can't detect array element length")
		}
		return r.Len() * elemLength, nil
	case reflect.String:
		if tagInfo.Length == 0 {
			return 0, errors.New("need to specify length")
		}
		return tagInfo.Length, nil
	}
	return getTypeBytesLength(r)
}
