package utils

import (
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"
	"bytes"
)

func HttpRequestWrap(uri string, targat func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				logger.Info("===================http server panic recover===============")
			}
		}()
		st := time.Now()
		logger.Debug("User-Agent: ", request.Header["User-Agent"])
		targat(response, request)
		logger.Debug(fmt.Sprintf("%s cost total time: %f ms", uri, time.Now().Sub(st).Seconds()*1000))
	}
}

func ReSettingLog() {
	// --------------------------------------------init log start
	logger.SetConsole(GlobalObject.SetToConsole)
	if GlobalObject.LogFileType == logger.ROLLINGFILE {
		logger.SetRollingFile(GlobalObject.LogPath, GlobalObject.LogName,
			GlobalObject.MaxLogNum, GlobalObject.MaxFileSize, GlobalObject.LogFileUnit)
	} else {
		logger.SetRollingDaily(GlobalObject.LogPath, GlobalObject.LogName)
		logger.SetLevel(GlobalObject.LogLevel)
	}
	// --------------------------------------------init log end
}

func XingoTry(f reflect.Value, args []reflect.Value, handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("-------------panic recover---------------")
			if handler != nil {
				handler(err)
			}
		}
	}()
	f.Call(args)
}

func Catch(f func())  {
	if err := recover(); err != nil {
		f()
	}
}

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