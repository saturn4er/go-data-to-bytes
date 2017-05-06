package d2b

import (
	"reflect"
	"strconv"
	"strings"

	"sync"

	"github.com/pkg/errors"
)

var structsTagsMx sync.RWMutex
var structsTags = make(map[reflect.Type][]*structFieldTag)

type structFieldTag struct {
	Length int
	Skip   bool
}

func parseStructFieldTag(field reflect.StructField) (*structFieldTag, error) {
	result := new(structFieldTag)
	tag := field.Tag.Get("d2b")
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "-" {
			result.Skip = true
			continue
		}
		if strings.HasPrefix(part, "length:") {
			sLength := strings.TrimPrefix(part, "length:")
			length, err := strconv.Atoi(sLength)
			if err != nil {
				return nil, err
			}
			result.Length = length
			continue
		}
	}
	return result, nil
}

func getStructTags(structType reflect.Type) ([]*structFieldTag, error) {
	structsTagsMx.RLock()
	if tags, ok := structsTags[structType]; ok {
		structsTagsMx.RUnlock()
		return tags, nil
	}
	structsTagsMx.RUnlock()
	structsTagsMx.Lock()
	defer structsTagsMx.Unlock()
	tags := make([]*structFieldTag, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		ft := structType.Field(i)
		tag, err := parseStructFieldTag(structType.Field(i))
		if err != nil {
			return nil, errors.Wrapf(err, "%v field tag error", ft.Name)
		}
		tags[i] = tag
	}
	structsTags[structType] = tags
	return structsTags[structType], nil
}
