package utils

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

func IsExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return IsExported(t.Name()) || t.PkgPath() == ""
}
