package Actor

import (
	"github.com/zllangct/RockGO/IMessage"
)

type Actor interface {
	Tell(message *ActorMessageInfo)error
	//GetActorID()ActorID
}



type IActorMessage interface {
	GetTitle()string
}

type IActorMessageHandler interface {
	IMessage.IMessageHanler
	MessageType()int
	MessageHandlerMap()map[string]func(message *ActorMessageInfo)
}