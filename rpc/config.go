package rpc

import "time"

/*
	set only effective before create client or server
*/
var (
	HeartInterval = time.Millisecond * 3000 //rpc heart beat interval
	CallTimeout   = time.Millisecond * 5000 //rpc call timeout
	Timeout       = time.Millisecond * 9000 //rpc timeout
	DebugMode     = true                    // is debug model
)
