package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

var ErrNodeOffline =errors.New("this node is offline")

type ActorProxyComponent struct {
	Component.Base
	nodeID        string
	localActors   sync.Map 			//本地actor [ActorID,*actor]
	nodeComponent *Cluster.NodeComponent
}

func (this *ActorProxyComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *ActorProxyComponent) IsUnique() bool {
	return true
}

func (this *ActorProxyComponent) Awake() {
	this.nodeID = Config.Config.ClusterConfig.LocalAddress
}

func (this *ActorProxyComponent) Register(actor *ActorComponent) error {
	actor.ActorID = NewActorID()
	id,err:= actor.ActorID.SetNodeID(this.nodeID)
	if err!=nil {
		return err
	}
	this.localActors.LoadOrStore(id, actor)
	return nil
}

func (this *ActorProxyComponent) Unregister(actor *ActorComponent) {
	if _, ok := this.localActors.Load(actor.ActorID); ok {
		return
	}
}



func (this *ActorProxyComponent) Emit(actorID ActorID, message IActorMessage) error {
	nodeID := actorID.GetNodeID()
	//本地消息不走网络
	if nodeID == this.nodeID {

		return nil
	}
	//非本地消息走网络代理
	client, err := this.nodeComponent.GetNodeClient(nodeID)
	if err!=nil {
		return err
	}

}

