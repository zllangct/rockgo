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

var ErrNodeOffline = errors.New("this node is offline")
var ErrNoThisActor = errors.New("no this actor")

type ActorProxyComponent struct {
	Component.Base
	nodeID        string
	localActors   sync.Map //本地actor [Target,*actor]
	nodeComponent *Cluster.NodeComponent
	location      *rpc.TcpClient
	isActorMode    bool
}

func (this *ActorProxyComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *ActorProxyComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *ActorProxyComponent) Awake() error{
	this.nodeID = Config.Config.ClusterConfig.LocalAddress
	this.isActorMode =Config.Config.ClusterConfig.IsActorModel
	err:= this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}
	//注册ActorProxyService服务
	s := new(ActorProxyService)
	s.init(this)
	err=this.nodeComponent.Register(s)
	if err != nil {
		return err
	}
	return nil
}

//通过角色获取一个actor
func (this *ActorProxyComponent)GetActorByRole(role string) (*ActorIDGroup,error) {
	location,err:=this.GetActorLocation()
	if err!=nil {
		return nil,err
	}
	var reply *ActorIDGroup
	err=location.Call("ActorLocationComponent.ServiceInquiry",role,&reply)
	if err!=nil {
		return nil, err
	}
	return reply,nil
}

//获取actor 位置服务器
func (this *ActorProxyComponent) GetActorLocation() (*rpc.TcpClient,error) {
	if this.location==nil {
		nodeID,err := this.nodeComponent.GetNode("actorlocation")
		if err!=nil {
			return nil,err
		}
		this.location,err=nodeID.GetClient()
		if err!=nil {
			return nil,err
		}
	}
	return this.location,nil
}
//注册actor服务
func (this *ActorProxyComponent) RoleRegister(role string,actor IActor) error {
	location,err:=this.GetActorLocation()
	if err!=nil {
		return err
	}
	var reply bool
	args:= ActorService{
		Role:role,
		ActorID:actor.ID(),
	}
	err=location.Call("ActorLocationComponent.ServiceRegister",args,&reply)
	if err!=nil {
		return err
	}
	return nil
}

//注册本地actor
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
//注销本地actor
func (this *ActorProxyComponent) Unregister(actor IActor) {
	if _, ok := this.localActors.Load(actor.ID()); ok {
		return
	}
}

//发送本地消息
func (this *ActorProxyComponent) LocalTell(actorID ActorID, messageInfo *ActorMessageInfo) error {
	v, ok := this.localActors.Load(actorID[2])
	if !ok {
		return ErrNoThisActor
	}
	actor, ok := v.(IActor)
	if !ok {
		return ErrNoThisActor
	}
	return actor.Tell(messageInfo.Sender,messageInfo.Message,messageInfo.reply)
}

//通过actor id 发送消息
func (this *ActorProxyComponent) Emit(actorID ActorID, messageInfo *ActorMessageInfo) error {
	logger.Debug(fmt.Sprintf("Actor: [ %s ] send message [ %s ] to actor [ %s ]",messageInfo.Sender.ID(),messageInfo.Message.Tittle,actorID.String()))
	nodeID := actorID.GetNodeID()
	//本地消息不走网络
	if nodeID == this.nodeID {
		return this.LocalTell(actorID,messageInfo)
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
