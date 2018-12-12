package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
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
	queueReceive chan *ActorMessageInfo //接收消息队列
	close        chan bool              //关闭信号
	active       int32                  //是否激活,0：未激活 1：激活
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

func (this *ActorComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_LOCAL
}

func (this *ActorComponent) Awake() {
	this.queueReceive= make(chan *ActorMessageInfo, 10)
	this.close=       make(chan bool)
	//初始化Actor代理
	err := this.Parent.Runtime().Root().Find(&this.Proxy)
	if err != nil {
		panic(err)
	}
	//初始化ID
	this.ActorID= EmptyActorID()
	//注册Actor到ActorProxy
	err = this.Proxy.Register(this)
	if err!=nil {
		logger.Error(err)
	}
	//设置Actor状态为激活
	atomic.StoreInt32(&this.active, 1)
	//初始化消息分发器
	go this.dispatch()
}

func (this *ActorComponent) Destroy() {
	this.close <- true
	//在ActorProxy取消注册
	this.Proxy.Unregister(this)
}

func (this *ActorComponent) Tell(messageInfo *ActorMessageInfo,reply ...*ActorMessage) error {
	if len(reply)>0 {
		messageInfo.Reply = reply[0]
	}
	if atomic.LoadInt32(&this.active) != 0 {
		this.queueReceive <- messageInfo
	} else {
		return errors.New("this actor is inactive or destroyed")
	}
	return nil
}

func (this *ActorComponent)Emit()  {

}

func (this *ActorComponent) ID() ActorID{
	return this.ActorID
}

func (this *ActorComponent) dispatch() {
	var messageInfo *ActorMessageInfo
	var ok bool
	handle := func(messageInfo *ActorMessageInfo) {
		cps := this.Parent.AllComponents()
		var err error = nil
		var val interface{}
		for val, err = cps.Next(); err == nil; val, err = cps.Next() {
			if messageHandler, ok := val.(IActorMessageHandler); ok {
				if handler, ok := messageHandler.MessageHandlers()[messageInfo.Message.Tittle]; ok {
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
