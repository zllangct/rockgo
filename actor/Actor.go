package Actor



type IActor interface {
	Tell(message *ActorMessageInfo,reply ...*ActorMessage) error
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

func (this *ActorRemote)Tell(messageInfo *ActorMessageInfo,reply ...*ActorMessage) error{
	if len(reply)==0 {
		r := new(ActorMessage)
		return this.proxy.Emit(this.ID(),messageInfo,r)
	}else{
		r:=reply[0]
		return this.proxy.Emit(this.ID(),messageInfo,r)
	}
}