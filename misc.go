package mysql

import (
	"reflect"
	"strings"
)

func camelToSnake(s string) string {
	var sb strings.Builder
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				sb.WriteByte('_')
			}
			sb.WriteByte(byte(c + 32))
		} else {
			sb.WriteByte(byte(c))
		}
	}
	return sb.String()
}

func tryGetFieldNameFromTag(tag reflect.StructTag) string {
	if tag, ok := tag.Lookup("db"); ok {
		return tag
	}
	return tryGetFieldNameFromOtherTag(tag)
}

func tryGetFieldNameFromOtherTag(tag reflect.StructTag) string {
	if tag, ok := tag.Lookup("json"); ok {
		return tag
	}
	if tag, ok := tag.Lookup("yaml"); ok {
		return tag
	}
	return ""
}
