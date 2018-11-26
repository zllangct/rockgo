package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/Component"
	"reflect"
)

/*
	Actor IComponent
	actor内线程安全
*/

type ActorComponent struct {
	Component.Base
	ActorID		 ActorID
	Proxy        *ActorProxy
	ActorType    ActorType              //Actor类型
	queueReceive chan *ActorMessageInfo //接收消息队列
	close        chan bool              //关闭信号
	active       bool
}

func NewActorComponent() *ActorComponent {
	return &ActorComponent{
		queueReceive:make(chan *ActorMessageInfo,10),
		close:make(chan bool),
	}
}



func (this *ActorComponent) IsUnique() bool  {
	return true
}

func (this *ActorComponent) Awake() {
	//初始化Actor代理
	err := this.Parent.Runtime().Root().Find(this.Proxy)
	if err != nil {
		panic(err)
	}
	//TODO 注册Actor到ActorProxy
	this.Proxy.Register(this)
	//初始化消息分发器
	go this.dispatch()
	this.active=true
}

func (this *ActorComponent)Destroy()  {
	//TODO 在ActorProxy取消注册
	this.Proxy.Unregister(this)
}

func (this *ActorComponent) Tell(message *ActorMessageInfo) error {
	message.Session.Self=this
	message.Session.SelfActorID=this.ActorID
	if this.active {
		this.queueReceive <- message
	}else{
		return errors.New("This actor is inactive or destroyed.")
	}
	return nil
}

func (this *ActorComponent) dispatch() {
	for {
		select {
		case <-this.close:
			this.active=false
			close(this.queueReceive)
			return
		case messageInfo := <-this.queueReceive:
			targetType := reflect.TypeOf((*IActorMessageHandler)(nil)).Elem()
			components := this.Parent.GetComponents(targetType)
			for val, err := components.Next(); err == nil; val, err = components.Next() {
				val.(IActorMessageHandler).MessageHandlerMap()[messageInfo.Message.GetTitle()](messageInfo)
			}
		}
	}
}
