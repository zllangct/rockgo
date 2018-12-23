package utils

import "time"

//条件等待
func When(interval time.Duration,conditions ...func()bool)  {
	var pass bool
	for{
		pass =true
		for _, cond := range conditions {
			if !cond() {
				pass =false
				break
			}
		}
		if pass {
			break
		}
		time.Sleep(interval)
	}
}
