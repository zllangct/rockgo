package main

import (
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
	"reflect"
)

type TestMessage struct {
	Name string
}

var Testid2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}):1,
}

type TestApi struct {
	network.Base
}

func newTestApi() *TestApi  {
	r:=&TestApi{}
	r.Base.Init(Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}