package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync/atomic"
)

/*
	actor IComponent
	actor内线程安全
*/
const(
	ACTOR_TYPE_LOCAL = ActorType(1)
	ACTOR_TYPE_REMOTE = ActorType(2)
)

type ActorType int

type ActorComponent struct {
	Component.Base
	ActorID      ActorID                //Actor地址
	Proxy        *ActorProxyComponent   //Actor代理
	ActorType    ActorType              //Actor类型
	queueReceive chan *ActorMessageInfo //接收消息队列
	close        chan bool              //关闭信号
	active       int32                  //是否激活,0：未激活 1：激活
}

func NewActorComponent() *ActorComponent {
	return &ActorComponent{
		queueReceive: make(chan *ActorMessageInfo, 10),
		close:        make(chan bool),
		ActorType:    ACTOR_TYPE_REMOTE, //默认为远程actor，内部逻辑使用，无跨节点通讯时，可设为本地，节约查询成本
	}
}

func (this *ActorComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&ActorProxyComponent{}),
	}
	return requires
}

func (this *ActorComponent) IsUnique() bool {
	return true
}

func (this *ActorComponent) Awake() {
	//初始化消息分发器
	go this.dispatch()
}

func (this *ActorComponent)Start(ctx *Component.Context)  {
	//初始化Actor代理
	err := this.Parent.Runtime().Root().Find(&this.Proxy)
	if err != nil {
		panic(err)
	}
	//注册Actor到ActorProxy
	this.Proxy.Register(this)
	//设置Actor状态为激活
	atomic.StoreInt32(&this.active, 1)
}

func (this *ActorComponent) Destroy() {
	this.close <- true
	//在ActorProxy取消注册
	this.Proxy.Unregister(this)
}

func (this *ActorComponent) Tell(message *ActorMessageInfo) error {
	message.Session.Self = this
	message.Session.SelfActorID = this.ActorID
	if atomic.LoadInt32(&this.active) != 0 {
		this.queueReceive <- message
	} else {
		return errors.New("this actor is inactive or destroyed")
	}
	return nil
}

func (this *ActorComponent)Emit()  {

}

func (this *ActorComponent) GetActorID() ActorID{
	return this.ActorID
}

func (this *ActorComponent) dispatch() {
	var messageInfo *ActorMessageInfo
	var ok bool
	handle := func(messageInfo *ActorMessageInfo) {
		infos := this.Parent.AllComponents()
		for _, info := range infos {
			if messageHandler, ok := info.Component.(IActorMessageHandler); ok {
				if handler, ok := messageHandler.MessageHandlers()[messageInfo.Message.GetTitle()]; ok {
					handler(messageInfo)
				}
			}
		}
	}
	for {
		select {
		case <-this.close:
			atomic.StoreInt32(&this.active, 0)
			close(this.queueReceive)
			//收到关闭信号后会继续处理完剩余消息
		case messageInfo, ok = <-this.queueReceive:
			if !ok {
				return
			}
			handle(messageInfo)
		}
	}
}
