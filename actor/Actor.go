package Actor



type IActor interface {
	Tell(message *ActorMessageInfo) error
	ID() ActorID
}

type IActorMessageHandler interface {
	MessageHandlers()map[string]func(message *ActorMessageInfo)
}

type ActorRemote struct {
	actorID ActorID
	proxy *ActorProxyComponent
}

func (this *ActorRemote)ID ()ActorID{
	return this.actorID
}

func (this *ActorRemote)Tell(messageInfo *ActorMessageInfo) error{
	return this.proxy.Emit(this.ID(),messageInfo)
}