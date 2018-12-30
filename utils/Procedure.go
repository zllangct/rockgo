package utils

import (
	"errors"
	"github.com/zllangct/RockGO/logger"
	"runtime/debug"
	"time"
)

/*  流程执行  */

type  Procedure struct {
	Task func()
	Condition func()bool
}

func StartProcedure(checkInterval time.Duration,tasks ...*Procedure)  {
	for _, task := range tasks {
		When(checkInterval,task.Condition)
		Try(task.Task)
	}
}

func Try(task func())  {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str =r.(error).Error()
			case string:
				str = r.(string)
			}
			logger.Error(errors.New(str+"\n"+ string(debug.Stack())))
		}
	})()
	task()
}