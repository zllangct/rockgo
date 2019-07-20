package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/timer"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"
)

/*
	actor component
	添加了ActorConponent组件的object可视为actor，该object上挂载的所有其他Component都可以响应actor消息
	默认状态下actor type 为 ACTOR_TYPE_ASYNC，在actor内是非线程安全的，需要有保证线程安全的措施
	可设置type为 ACTOR_TYPE_SYNC 此时所有消息穿行化，actor内线程安全
*/
const (
	ACTOR_TYPE_DEFAULT ActorType = iota
	ACTOR_TYPE_SYNC
	ACTOR_TYPE_ASYNC
)

type ActorType int

type ActorComponent struct {
	Component.Base
	ActorType    ActorType
	ActorID      ActorID                //Actor地址
	Proxy        *ActorProxyComponent   //Actor代理
	queueReceive chan *ActorMessageInfo //接收消息队列
	close        chan bool              //关闭信号
	active       int32                  //是否激活,0：未激活 1：激活
}

func NewActorComponent(actorType ActorType) *ActorComponent {
	return &ActorComponent{ActorType: actorType}
}

func (this *ActorComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Runtime().Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&ActorProxyComponent{}),
	}
	return requires
}

func (this *ActorComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_LOCAL
}

func (this *ActorComponent) Initialize() error {
	this.queueReceive = make(chan *ActorMessageInfo, 20)
	this.close = make(chan bool)
	//初始化actor类型
	if this.ActorType == ACTOR_TYPE_DEFAULT {
		this.ActorType = ACTOR_TYPE_ASYNC
	}
	//初始化Actor代理
	err := this.Runtime().Root().Find(&this.Proxy)
	if err != nil {
		logger.Error(err)
		return err
	}
	//初始化ID
	this.ActorID = EmptyActorID()
	//注册Actor到ActorProxy
	err = this.Proxy.Register(this)
	if err != nil {
		logger.Error(err)
		return err
	}
	//初始化消息分发器
	go this.dispatch()
	//设置Actor状态为激活
	atomic.StoreInt32(&this.active, 1)
	return nil
}

func (this *ActorComponent) RegisterService(service string) error {
	return this.Proxy.RegisterService(this, service)
}
func (this *ActorComponent) UnregisterService(service string) {
	this.Proxy.UnregisterService(service)
}

func (this *ActorComponent) Destroy(ctx *Component.Context) {
	this.close <- true
	//在ActorProxy取消注册
	this.Proxy.Unregister(this)
}

func (this *ActorComponent) Tell(sender IActor, message *ActorMessage, reply ...**ActorMessage) error {
	if atomic.LoadInt32(&this.active) == 0 {
		return errors.New("this actor is inactive or destroyed")
	}

	messageInfo := &ActorMessageInfo{
		Sender:  sender,
		Message: message,
	}

	if len(reply) > 0 {
		messageInfo.NeedReply(true)
		messageInfo.reply = reply[0]
	} else {
		messageInfo.NeedReply(false)
	}

	this.queueReceive <- messageInfo

	if messageInfo.IsNeedReply() {
		select {
		case <-timer.After(time.Duration(Config.Config.ClusterConfig.RpcCallTimeout) * time.Millisecond):
			messageInfo.err = ErrTimeout
		case <-messageInfo.done:
		}
	}
	return messageInfo.err
}

func (this *ActorComponent) Emit() {

}

func (this *ActorComponent) ID() ActorID {
	return this.ActorID
}

func (this *ActorComponent) dispatch() {
	var messageInfo *ActorMessageInfo
	var ok bool
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
			switch this.ActorType {
			case ACTOR_TYPE_SYNC:
				this.handle(messageInfo)
			case ACTOR_TYPE_ASYNC:
				go this.handle(messageInfo)
			default:

			}
		}
	}
}

func (this *ActorComponent) handle(messageInfo *ActorMessageInfo) {
	cps := this.Parent().AllComponents()
	var err error = nil
	var val interface{}
	for val, err = cps.Next(); err == nil; val, err = cps.Next() {
		if messageHandler, ok := val.(IActorMessageHandler); ok {
			if handler, ok := messageHandler.MessageHandlers()[messageInfo.Message.Service]; ok {
				this.Catch(handler, messageInfo)
			}
		}
	}
}

func (this *ActorComponent) Catch(handler func(message *ActorMessageInfo) error, m *ActorMessageInfo) {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str + string(debug.Stack()))
			logger.Error(err)
			m.replyError(err)
		}
	})()
	err := handler(m)
	if err != nil {
		m.replyError(err)
	} else {
		m.replySuccess()
	}
}
