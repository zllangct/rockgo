package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"reflect"
	"runtime/debug"
	"sync/atomic"
)

/*
	async actor IComponent
	异步任务actor内 非线程安全
*/

type ActorAsyncComponent struct {
	Component.Base
	ActorID      ActorID                //Actor地址
	Proxy        *ActorProxyComponent   //Actor代理
	Role         string
	close        chan bool              //关闭信号
	active       int32                  //是否激活,0：未激活 1：激活
}

func (this *ActorAsyncComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	//添加该组件需要根节点拥有ActorProxyComponent,ConfigComponent组件
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&ActorProxyComponent{}),
	}
	return requires
}

func (this *ActorAsyncComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_LOCAL
}

func (this *ActorAsyncComponent) Awake()error {
	this.close=       make(chan bool)
	//初始化Actor代理
	err := this.Parent.Runtime().Root().Find(&this.Proxy)
	if err != nil {
		return err
	}
	//初始化ID
	this.ActorID= EmptyActorID()
	//注册Actor到ActorProxy
	err = this.Proxy.Register(this)
	if err!=nil {
		return err
	}
	//设置Actor状态为激活
	atomic.StoreInt32(&this.active, 1)
	return nil
}

func (this *ActorAsyncComponent) Destroy() {
	this.close <- true
	//在ActorProxy取消注册
	this.Proxy.Unregister(this)
}

func (this *ActorAsyncComponent) Tell(sender IActor,message *ActorMessage,reply ...**ActorMessage) error {
	if atomic.LoadInt32(&this.active) != 0 {
		return errors.New("this actor is inactive or destroyed")
	}

	messageInfo:=&ActorMessageInfo{
		Sender:sender,
		Message:message,
	}

	if len(reply)>0 {
		messageInfo.NeedReply(true)
		messageInfo.reply = reply[0]
	}else{
		messageInfo.NeedReply(false)
	}

	go this.handle(messageInfo)

	if messageInfo.IsNeedReply() {
		<-messageInfo.done
	}
	return messageInfo.err
}

func (this *ActorAsyncComponent)Emit()  {

}

func (this *ActorAsyncComponent) ID() ActorID{
	return this.ActorID
}

func (this *ActorAsyncComponent) handle(messageInfo *ActorMessageInfo) {
	cps := this.Parent.AllComponents()
	var err error = nil
	var val interface{}
	for val, err = cps.Next(); err == nil; val, err = cps.Next() {
		if messageHandler, ok := val.(IActorMessageHandler); ok {
			if handler, ok := messageHandler.MessageHandlers()[messageInfo.Message.Tittle]; ok {
				this.Catch(handler,messageInfo)
			}
		}
	}
}

func (this *ActorAsyncComponent) Catch(handler func(message *ActorMessageInfo),m *ActorMessageInfo) {
	defer (func() {
		if r := recover(); r != nil {
			err := errors.New(r.(error).Error() + "\n" + string(debug.Stack()))
			logger.Error(err)
			m.CallError(err)
		}
	})()
	handler(m)
}