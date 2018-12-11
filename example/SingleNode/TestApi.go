package main

import (
	"fmt"
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
)



//协议接口组
type TestApi struct {
	network.Base
}

func NewTestApi() *TestApi  {
	r:=&TestApi{}
	r.Init(r,Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}

func (this *TestApi)Hello(sess *network.Session,message *TestMessage)  {
	println(fmt.Sprintf("Hello,%s",message.Name))
	println(fmt.Sprintf("this api parent:%s",this.GetParent().Name()))
}