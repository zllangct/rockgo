package utils

import (
	"bytes"
	"errors"
	"github.com/zllangct/RockGO/logger"
	"reflect"
	"runtime/debug"
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

func Try(task func(), catch ...func(error)) {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str + "\n" + string(debug.Stack()))
			if len(catch) > 0 {
				catch[0](err)
			} else {
				logger.Error(err)
			}
		}
	})()
	task()
}
