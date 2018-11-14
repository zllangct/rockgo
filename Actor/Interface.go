package Actor

import (
	"github.com/zllangct/RockGO/IMessage"
	"github.com/zllangct/RockGO/Component"
)

type Actor interface {
	Tell(message *ActorMessageInfo)
	Emit(message *ActorMessageInfo)
}

type IActorMessage interface {
	//IMessage.IMessage
	GetTitle()string
	GetRecipientsID()int64
}

type IActorMessageHandler interface {
	IMessage.IMessageHanler
	MessageType()int
	MessageHandlerMap()map[string]func(obj *Component.Object,message *ActorMessageInfo)
}