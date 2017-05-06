package d2b

import (
	"encoding/binary"
	"math"
	"reflect"

	"github.com/pkg/errors"
)

// ConvertBytesToData write byte array to data
// Can panic, with slice bounds out of range if there's not enough bytes
func Decode(bytes []byte, endian binary.ByteOrder, data interface{}) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Ptr {
		return errors.New("data should be pointer")
	}
	v := reflect.ValueOf(data)
	if v.IsNil() {
		return errors.New("can't decode to nil pointer")
	}
	_, err := updateValueByTypeFromBytess(v.Elem(), bytes, endian)
	return err
}

func updateValueByTypeFromBytess(v reflect.Value, bytes []byte, endian binary.ByteOrder) ([]byte, error) {
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return updateValueByTypeFromBytess(v.Elem(), bytes, endian)
	case reflect.Int8:
		v.SetInt(int64(int8(bytes[0])))
		return bytes[1:], nil
	case reflect.Int16:
		val := endian.Uint16(bytes[:2])
		v.SetInt(int64(int16(val)))
		return bytes[2:], nil
	case reflect.Int32:
		val := endian.Uint32(bytes[:4])
		v.SetInt(int64(int32(val)))
		return bytes[4:], nil
	case reflect.Int64:
		val := endian.Uint64(bytes[:8])
		v.SetInt(int64(val))
		return bytes[8:], nil
	case reflect.Uint8:
		v.SetUint(uint64(bytes[0]))
		return bytes[1:], nil
	case reflect.Uint16:
		val := endian.Uint16(bytes[:2])
		v.SetUint(uint64(int16(val)))
		return bytes[2:], nil
	case reflect.Uint32:
		val := endian.Uint32(bytes[:4])
		v.SetUint(uint64(int16(val)))
		return bytes[4:], nil
	case reflect.Uint64:
		v.SetUint(endian.Uint64(bytes[:8]))
		return bytes[8:], nil
	case reflect.Float32:
		val := endian.Uint32(bytes[:4])
		float := math.Float32frombits(val)
		v.SetFloat(float64(float))
		return bytes[4:], nil
	case reflect.Float64:
		val := endian.Uint64(bytes[:8])
		float := math.Float64frombits(val)
		v.SetFloat(float)
		return bytes[8:], nil
	case reflect.Array:
		var err error
		for i := 0; i < v.Len(); i++ {
			bytes, err = updateValueByTypeFromBytess(v.Index(i), bytes, endian)
			if err != nil {
				return []byte{}, err
			}
		}
		return bytes, nil
	case reflect.Struct:
		tags, err := getStructTags(t)
		if err != nil {
			return []byte{}, errors.Wrap(err, "can't parse struct tags")
		}
		for i := 0; i < t.NumField(); i++ {
			fv := v.Field(i)
			bytes, err = updateStructField(fv, bytes, tags[i], endian)
			if err != nil {
				ft := t.Field(i)
				return []byte{}, errors.Wrapf(err, "can't update struct field %s.%s", t.Name(), ft.Name)
			}
		}
		return bytes, nil
	default:
		return []byte{}, errors.Errorf("type %v is not supported", t.Kind())
	}
}

func updateStructField(v reflect.Value, bytes []byte, tags *structFieldTag, endian binary.ByteOrder) ([]byte, error) {
	if tags.Skip {
		return bytes, nil
	}
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return updateStructField(v.Elem(), bytes, tags, endian)
	case reflect.Slice:
		if tags.Length == 0 {
			return nil, errors.New("empty length")
		}
		var err error
		for i := 0; i < v.Len(); i++ {
			bytes, err = updateValueByTypeFromBytess(v.Index(i), bytes, endian)
			if err != nil {
				return []byte{}, err
			}
		}
		l := v.Len()
		for i := 0; i < tags.Length-l; i++ {
			value := reflect.New(t.Elem())
			bytes, err = updateValueByTypeFromBytess(value, bytes, endian)
			if err != nil {
				return []byte{}, err
			}
			v.Set(reflect.Append(v, value.Elem()))
		}
		return bytes, nil
	case reflect.String:
		if tags.Length == 0 {
			return nil, errors.New("empty length")
		}
		v.SetString(bytesToStr(bytes[:tags.Length]))
		return bytes[tags.Length:], nil
	}
	return updateValueByTypeFromBytess(v, bytes, endian)
}
