package main

import (
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
)



//协议接口组
type TestApi struct {
	network.Base
}

func newTestApi() *TestApi  {
	r:=&TestApi{}
	r.Base.Init(r,Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}

func (this *TestApi)Hello()  {
	println("Hello RockGO")
}