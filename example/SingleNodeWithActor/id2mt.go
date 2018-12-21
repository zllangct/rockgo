package main

import "reflect"

/*
    此文件是服务端和客户端消息号对应，根据项目差异可以根据实际情况
	通过各种方式得到，建议通过工具统一生成，避免服务端和客户端不匹配。

*/
var Testid2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}):1,
	reflect.TypeOf(&TestCreateRoom{}):2,
	reflect.TypeOf(&CreateResult{}):3,
}