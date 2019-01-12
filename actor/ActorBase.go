package Actor

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/logger"
)

type ActorBase struct {
	MessageHandler map[string]func(message *ActorMessageInfo)
	Actor   *ActorComponent
	parent  *Component.Object
}

func (this *ActorBase)ActorInit(parent *Component.Object)  {
	this.parent = parent
}

func (this *ActorBase) MessageHandlers() map[string]func(message *ActorMessageInfo) {
	return this.MessageHandler
}

func (this *ActorBase)AddHandler(service string,handler func(message *ActorMessageInfo),isService ...bool)  {
	if this.MessageHandler == nil{
		this.MessageHandler = map[string]func(message *ActorMessageInfo){}
	}
	this.MessageHandler[service]=handler
	if this.Actor!=nil && len(isService)>0{
		err:=this.Actor.RegisterService(service)
		if err!=nil {
			logger.Error(err)
		}
	}
}