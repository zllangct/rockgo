package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/rpc"
	"math/rand"
	"net"
	"reflect"
	"sync"
)

var ErrNodeOffline =errors.New("this node is offline")

type ActorProxyComponent struct {
	Component.Base
	nodeName         string
	localActors    sync.Map 			//本地actor [ActorID,*actor]
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
	//oMaster,err:= this.Parent.Root().FindObject("master")
	//if err!=nil {
	//	logger.Fatal("get master object failed")
	//	panic(err)
	//	return
	//}
	//err= oMaster.Find(&this.)
	//if err != nil {
	//	logger.Fatal("get node component failed")
	//	panic(err)
	//	return
	//}
}

func (this *ActorProxyComponent) Register(actor *ActorComponent) {
	if _, ok := this.localActors.LoadOrStore(actor.ActorID, actor); ok {
		return
	}
}

func (this *ActorProxyComponent) Unregister(actor *ActorComponent) {
	if _, ok := this.localActors.Load(actor.ActorID); ok {
		return
	}
}



func (this *ActorProxyComponent) Emit(actorID ActorID, service string, args interface{}, reply interface{}) error {
	nodeName := actorID.GetNodeName()
	//本地消息不走网络
	if nodeName == this.nodeName {

		return nil
	}

	//if !this.isOnline {
	//	return ErrNodeOffline
	//}
	
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

//获取位置节点
func (this *ActorProxyComponent)GetLocationNodes()*rpc.TcpClient {
	this.locationMutex.Lock()
	defer this.locationMutex.Unlock()
	//已存在
	if len(this.locationClient) > 0 {
		rnd:=rand.Intn(len(this.locationClient))
		return this.locationClient[rnd]
	}
	//不存在

}

//从位置服务节点查询
func (this *ActorProxyComponent)LocationInquiry()  {
	this.locationMutex.Lock()
	idx:=rand.Intn(len(this.locationClient))
	locationClient:=this.locationClient[idx]
	this.locationMutex.Unlock()
	err := locationClient.Call(service, args, reply)
}

//查询节点,从master上面查询
func (this *ActorProxyComponent)NodeInquiry(role string)  {
	reply:=make([]string,0)
	this.Parent.Root().Find()

}

