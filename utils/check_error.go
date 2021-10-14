package utils

import (
	"errors"
	"github.com/zllangct/rockgo/logger"
	"runtime/debug"
)

func CheckError() {
	if r := recover(); r != nil {
		var str string
		switch r.(type) {
		case error:
			str = r.(error).Error()
		case string:
			str = r.(string)
		}
		err := errors.New("\n" + str + "\n" + string(debug.Stack()))
		logger.Error(err)
	}
}
