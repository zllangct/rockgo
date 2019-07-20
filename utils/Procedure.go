package utils

import (
	"time"
)

/*  流程执行  */

type Procedure struct {
	Task      func()
	Condition func() bool
}

func StartProcedure(checkInterval time.Duration, tasks ...*Procedure) {
	for _, task := range tasks {
		When(checkInterval, task.Condition)
		Try(task.Task)
	}
}
