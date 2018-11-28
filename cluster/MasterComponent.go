package Cluster

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type MasterComponent struct {
	Component.Base
	Nodes  sync.Map
}

func (this *MasterComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func(this *MasterComponent)Awake(){

}

func (this *MasterComponent)MessageHandlers()map[string]func(message *Actor.ActorMessageInfo)  {

}
