package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/utils/UUID"
	"reflect"
	"sync"
)

var ErrNodeOffline = errors.New("this node is offline")
var ErrNoThisActor = errors.New("no this actor")

type ActorProxyComponent struct {
	Component.Base
	nodeID        string
	localActors   sync.Map //本地actor [Target,*actor]
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

func (this *ActorProxyComponent) Register(actor IActor) error {
	id:=actor.ID()
	id[2]=UUID.Next()
	id, err := id.SetNodeID(this.nodeID)
	if err != nil {
		return err
	}
	this.localActors.LoadOrStore(id[2], actor)
	return nil
}

func (this *ActorProxyComponent) Unregister(actor IActor) {
	if _, ok := this.localActors.Load(actor.ID()); ok {
		return
	}
}

//本地消息
func (this *ActorProxyComponent) LocalTell(actorID ActorID, messageInfo *ActorMessageInfo) error {
	v, ok := this.localActors.Load(actorID)
	if !ok {
		return ErrNoThisActor
	}
	actor, ok := v.(IActor)
	if !ok {
		return ErrNoThisActor
	}
	return actor.Tell(messageInfo)
}

func (this *ActorProxyComponent) Emit(actorID ActorID, messageInfo *ActorMessageInfo) error {
	nodeID := actorID.GetNodeID()
	//本地消息不走网络
	if nodeID == this.nodeID {

		return nil
	}
	//非本地消息走网络代理
	client, err := this.nodeComponent.GetNodeClient(nodeID)
	if err != nil {
		return err
	}
	var reply = false
	err = client.Call("ActorService.Tell", &ActorRpcMessageInfo{
		Target:  actorID,
		Sender:  messageInfo.Sender.ID(),
		Message: messageInfo.Message,
	}, &reply)
	if err != nil {
		return err
	}
	return nil
}
