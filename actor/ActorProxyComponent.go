package Actor

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"net"
	"reflect"
	"sync"
)

var ErrNodeOffline =errors.New("this node is offline")

type ActorProxyComponent struct {
	Component.Base
	nodeName         string
	localActors    sync.Map 			//本地actor [ActorID,*actor]
	config         *Config.ConfigComponent
	nodeComponent  *Cluster.NodeComponent
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
	err := this.Parent.Root().Find(&this.config)
	if err != nil {
		logger.Fatal("get config component failed")
		panic(err)
		return
	}
}

func (this *ActorProxyComponent) Register(actor *ActorComponent) error {
	actor.ActorID = NewActorID()
	id,err:= actor.ActorID.SetNodeID(this.nodeName)
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



func (this *ActorProxyComponent) Emit(actorID ActorID, service string, args interface{}, reply interface{}) error {
	nodeName := actorID.GetNodeID()
	//本地消息不走网络
	if nodeName == this.nodeName {

		return nil
	}
	if !this.nodeComponent.IsOnline() {
		return ErrNodeOffline
	}
	//客户端已经存在
	clientInterface, ok := this.rpcClient.Load(id)
	if ok {
		client := clientInterface.(*rpc.TcpClient)
		err := client.Call(service, args, reply)
		switch err {
		case rpc.ErrConnClosing,rpc.ErrShutdown,rpc.ErrErrorBody,rpc.ErrTimeout:
			this.rpcClient.Delete(id)
		default:
			return err
		}
	}
	//客户端不存在,先在位置服务器上查询actor所在节点
	if len(this.locationClient)>0{
		MasterAddr := &net.TCPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: config.ClusterConfig.LocalPort,
			Zone: "",
		}

		return err
	}
	//如果位置服务器上没有，从master上面查询




}

