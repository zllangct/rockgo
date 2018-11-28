package Actor



type Actor interface {
	Tell(message *ActorMessageInfo)error
	//GetActorID()ActorID
}

type IActorMessage interface {
	GetTitle()string
}

type IActorMessageHandler interface {
	MessageType()int
	MessageHandlers()map[string]func(message *ActorMessageInfo)
}