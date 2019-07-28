package utils

import (
	"bytes"
	"sync"
)
 var pool  = sync.Pool{
 	New: func() interface{} {
		return &bytes.Buffer{}
	},
 }

func StrToBytes(strData string) []byte {
	buffer := pool.Get().(*bytes.Buffer)

	defer pool.Put(buffer)

	buffer.Reset()
	buffer.WriteString(strData)

	return buffer.Bytes()
}

func BytesToStr(b []byte) string {
	buffer := pool.Get().(*bytes.Buffer)

	defer pool.Put(buffer)

	buffer.Reset()
	buffer.Write(b)

	return buffer.String()
}