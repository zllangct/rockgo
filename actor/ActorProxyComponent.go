package Actor

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"github.com/zllangct/RockGO/utils/UUID"
	"reflect"
	"sync"
)

var ErrNoThisService = errors.New("no this service")
var ErrNoThisActor = errors.New("no this actor")

type ActorProxyComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeID        string
	localActors   sync.Map //本地actor [Target,actor]
	service       sync.Map // [service,[]actor]
	nodeComponent *Cluster.NodeComponent
	location      *rpc.TcpClient
	//isActorMode   bool
	isOnline      bool
}

func (this *ActorProxyComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Runtime().Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *ActorProxyComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *ActorProxyComponent) Initialize() error {
	logger.Info("ActorProxyComponent init .....")
	this.nodeID = Config.Config.ClusterConfig.LocalAddress
	//this.isActorMode = Config.Config.ClusterConfig.IsActorModel
	err := this.Runtime().Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}
	//注册ActorProxyService服务
	s := new(ActorProxyService)
	s.init(this)
	err = this.nodeComponent.Register(s)
	if err != nil {
		return err
	}
	logger.Info("ActorProxyComponent initialized.")
	return nil
}

func (this *ActorProxyComponent) IsOnline() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.isOnline
}

func (this *ActorProxyComponent) Destroy(ctx *Component.Context)  {

}

//调用actor service
func (this *ActorProxyComponent) ServiceCall(actor IActor,message *ActorMessage, reply **ActorMessage, role ...string) error {
	_,err:=this.ServiceCallReturnClient(actor,message,reply,role...)
	return err
}

//调用actor service
func (this *ActorProxyComponent) ServiceCallReturnClient(actor IActor,message *ActorMessage, reply **ActorMessage, role ...string) (*rpc.TcpClient,error ){
	var targetID ActorID
	g, ok := this.service.Load(message.Service)
	//优先尝试本地服务
	if ok {
		targetID = g.(*ActorIDGroup).RndOne()
		messageInfo := &ActorMessageInfo{
			Sender:  actor,
			Message: message,
			reply:   reply,
		}
		err:= this.Emit(targetID, messageInfo)
		if err==nil {
			return nil,err
		}
	}
	//本地无此服务，且有role参数时走网络途径
	if len(role) == 0 {
		return nil,ErrNoThisService
	}
	nodeID, err := this.nodeComponent.GetNode(role[0])
	if err != nil {
		return nil,err
	}
	if nodeID.Addr == this.nodeID {
		return nil,ErrNoThisService
	}
	client, err := nodeID.GetClient()
	if err != nil {
		return nil,err
	}
	err =this.ServiceCallByRpcClient(actor,message,reply,client)
	if err != nil {
		return nil,err
	}

	return client, nil
}

//根据rpc客户端直接调用actor服务
func (this *ActorProxyComponent) ServiceCallByRpcClient(actor IActor,message *ActorMessage, reply **ActorMessage, client *rpc.TcpClient)error  {
	args := &ServiceCall{
		Sender:  actor.ID(),
		Message: message,
	}
	err := client.Call("ActorProxyService.ServiceCall", args, reply)
	if err != nil {
		return err
	}
	return nil
}

//注册服务
func (this *ActorProxyComponent) RegisterServiceUnique(actor IActor, service string) error {
	g, _ := this.service.LoadOrStore(service, &ActorIDGroup{})
	if !g.(*ActorIDGroup).Has(actor.ID()) {
		g.(*ActorIDGroup).Add(actor.ID())
	}else{
		return errors.New("this service is repeated")
	}
	return nil
}
func (this *ActorProxyComponent) RegisterServices(actor IActor, service ...string) error {
	for _, value := range service {
		g, _ := this.service.LoadOrStore(value, &ActorIDGroup{})
		if !g.(*ActorIDGroup).Has(actor.ID()) {
			g.(*ActorIDGroup).Add(actor.ID())
		}
	}
	return nil
}

//取消注册服务
func (this *ActorProxyComponent) UnregisterService(actor IActor, service ...string) error {
	for _, value := range service {
		g, ok := this.service.Load(value)
		if ok && !g.(*ActorIDGroup).Has(actor.ID()) {
			g.(*ActorIDGroup).Sub(actor.ID())
		}
	}
	return nil
}

//注册本地actor
func (this *ActorProxyComponent) Register(actor IActor) error {
	id := actor.ID()
	id[2] = UUID.Next()
	id, err := id.SetNodeID(this.nodeID)
	if err != nil {
		return err
	}
	this.localActors.Store(id.String(), actor)
	return nil
}

//注销本地actor
func (this *ActorProxyComponent) Unregister(actor IActor) {
	if _, ok := this.localActors.Load(actor.ID().String()); ok {
		this.localActors.Delete(actor.ID().String())
		return
	}
}

//发送本地消息
func (this *ActorProxyComponent) LocalTell(actorID ActorID, messageInfo *ActorMessageInfo) error {
	v, ok := this.localActors.Load(actorID.String())
	if !ok {
		return ErrNoThisActor
	}
	actor, ok := v.(IActor)
	if !ok {
		return ErrNoThisActor
	}
	return actor.Tell(messageInfo.Sender, messageInfo.Message, messageInfo.reply)
}

//通过actor id 发送消息
func (this *ActorProxyComponent) Emit(actorID ActorID, messageInfo *ActorMessageInfo) error {
	logger.Debug(fmt.Sprintf("Actor: [ %s ] send message [ %s ] to actor [ %s ]", messageInfo.Sender.ID().String(), messageInfo.Message.Service, actorID.String()))
	nodeID := actorID.GetNodeID()
	//本地消息不走网络
	if nodeID == this.nodeID {
		return this.LocalTell(actorID, messageInfo)
	}
	//非本地消息走网络代理
	client, err := this.nodeComponent.GetNodeClient(nodeID)
	if err != nil {
		return err
	}
	err = client.Call("ActorProxyService.Tell", &ActorRpcMessageInfo{
		Target:  actorID,
		Sender:  messageInfo.Sender.ID(),
		Message: messageInfo.Message,
	}, messageInfo.reply)
	if err != nil {
		return err
	}
	return nil
}
