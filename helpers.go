package dtb

import "reflect"

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