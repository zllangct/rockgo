package Actor

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type ActorProxyComponent struct {
	Component.Base
	localActors sync.Map //[ActorID,*actor]
	remoteActors sync.Map //[ActorID,*actor]
}

func (this *ActorProxyComponent) GetRequire() (requires map[*Component.Object][]reflect.Type) {
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return
}

func (this *ActorProxyComponent) IsUnique() bool {
	return true
}

func (thih *ActorProxyComponent)Awake()  {

}

func (this *ActorProxyComponent)Register(actor *ActorComponent)  {
	if _,ok:=this.localActors.LoadOrStore(actor.ActorID,actor) ;ok{
		return
	}
	//如果是远程Actor,则注册ActorID到分布式
	if actor.ActorType == ACTOR_TYPE_REMOTE {

	}
}

func (this *ActorProxyComponent)Unregister(actor *ActorComponent)  {
	if _,ok:=this.localActors.Load(actor.ActorID) ;ok{
		//如果是远程Actor,从分布式上取消注册
		if actor.ActorType == ACTOR_TYPE_REMOTE {

		}
		return
	}
}