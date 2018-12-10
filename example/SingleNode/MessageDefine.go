package main
/*
	消息结构体定义，定义可以在任意可访问的地方，便于统一管理，建议
	统一在统一文件，或按照一定原则分组。也可由protobuf的工具生成。
*/

type TestMessage struct {
	Name string
}

