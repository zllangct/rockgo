package Actor

import (
	"github.com/zllangct/RockGO/Component"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
)

/*
	Actor Component
*/

type ActorComponent struct {
	obj          *Component.Object //父对象
	ActorType    ActorType         //Actor类型
	queueReceive *utils.SyncQueue  //接收消息队列
	close        chan bool         //关闭信号
}

func (this *ActorComponent) New() Component.Component {
	return &ActorComponent{
		ActorType: ActorType(1),
	}
}

func (this *ActorComponent) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this *ActorComponent) Awake(parent *Component.Object) {
	this.obj = parent
	go this.Dispatch()
}

func (this *ActorComponent) Start(ctx *Component.Context) {
	//如果是远程Actor,则注册ActorID到分布式  //TODO 需要在actor销毁时在分布式上取消注册
	if this.ActorType != 0 && this.ActorType == ACTOR_TYPE_REMOTE {

	}
}

func (this *ActorComponent) Update(ctx *Component.Context) {

}

func (this *ActorComponent) Serialize() (interface{}, error) {
	return nil, nil
}

func (this *ActorComponent) Deserialize(raw interface{}) error {
	return nil
}

func (this *ActorComponent) Tell(message *ActorMessageInfo) {
	message.Recipients = this
	this.queueReceive.Push(message)
}

func (this *ActorComponent) Emit(message *ActorMessageInfo) {
	// Checking the message package, title and recipients are required
	if message.Message.GetTitle() == "" || message.Message.GetRecipientsID() == 0 {
		logger.Fatal("Checking the message package, title and recipients are required")
		return
	}
	if message.Recipients != nil {
		message.Recipients.Tell(message)
	} else if message.Session != nil {
		//TODO 走代理层
	} else {
		//TODO 通过ActorID 判断途径
	}
}

func (this *ActorComponent) Dispatch() {
	for {
		select {
		case <-this.close:
		case messageInfo := this.queueReceive.Pop().(*ActorMessageInfo):
			components := this.obj.GetComponents(reflect.TypeOf((IActorMessageHandler)(nil)))
			for val, err := components.Next(); err == nil; val, err = components.Next() {
				val.(IActorMessageHandler).MessageHandlerMap()[messageInfo.Message.GetTitle()](this.obj, messageInfo)
			}
		}
	}
}
