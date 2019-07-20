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
	isOnline bool
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

func (this *ActorProxyComponent) Destroy(ctx *Component.Context) {

}

//获取本地actor服务
func (this *ActorProxyComponent) GetLocalActorService(serviceName string) (*ActorService, error) {
	var service *ActorService
	var err error
	s, ok := this.service.Load(serviceName)
	if !ok {
		return nil, ErrNoThisService
	}
	service = s.(*ActorService)
	if err != nil {
		return nil, err
	}
	return service, nil
}

//获取actor服务
func (this *ActorProxyComponent) GetActorService(role string, serviceName string) (*ActorService, error) {
	var service *ActorService
	var err error
	//优先尝试本地服务
	service, err = this.GetLocalActorService(serviceName)
	if err == nil {
		return service, nil
	}

	//获取远程服务
	if role == LOCAL_SERVICE {
		return nil, errors.New("role is empty")
	}
	client, err := this.nodeComponent.GetNodeClientByRole(role)
	if err != nil {
		return nil, err
	}
	var reply ActorID
	err = client.Call("ActorProxyService.ServiceInquiry", service, &reply)
	if err != nil {
		return nil, err
	}
	return NewActorService(NewActor(reply, this), serviceName), nil
}

//注册服务
func (this *ActorProxyComponent) RegisterService(actor IActor, service string) error {
	_, ok := this.service.Load(service)
	if ok {
		return errors.New("this service is repeated")
	}
	this.service.Store(service, NewActorService(actor, service))
	return nil
}

//取消注册服务
func (this *ActorProxyComponent) UnregisterService(service string) {
	this.service.Delete(service)
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
	logger.Debug(fmt.Sprintf("actor: [ %s ] send message [ %s ] to actor [ %s ]", messageInfo.Sender.ID().String(), messageInfo.Message.Service, actorID.String()))
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
