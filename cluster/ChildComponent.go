package Cluster

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type ChildComponent struct {
	Component.Base
	Nodes  sync.Map
}

func (this *ChildComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func(this *ChildComponent)Awake(){

}

func (this *ChildComponent)MessageHandlers()map[string]func(message *Actor.ActorMessageInfo)  {

}
