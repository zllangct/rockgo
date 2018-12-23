package utils

import (
	"bytes"
	"reflect"
	"unicode"
	"unicode/utf8"
)


func StrToBytes(strData string) []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteString(strData)
	return buffer.Bytes()
}

func BytesToStr(b []byte) string {
	buffer := &bytes.Buffer{}
	buffer.Write(b)
	return buffer.String()
}

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
