package Actor



type IActor interface {
	Tell(message *ActorMessageInfo)error
	GetActorID() ActorID
}

type IActorMessage interface {
	GetTitle()string
}

type IActorMessageHandler interface {
	MessageHandlers()map[string]func(message *ActorMessageInfo)
}