package Actor

type IActor interface {
	Tell(sender IActor,message *ActorMessage,reply ...**ActorMessage) error
	ID() ActorID
}

type IActorMessageHandler interface {
	MessageHandlers()map[string]func(message *ActorMessageInfo)error
}

type Actor struct {
	actorID ActorID
	proxy *ActorProxyComponent
}

func NewActor(id ActorID,proxy *ActorProxyComponent) IActor {
	return &Actor{
		actorID:id,
		proxy:proxy,
	}
}

func (this *Actor)ID ()ActorID{
	return this.actorID
}

func (this *Actor)Tell(sender IActor,message *ActorMessage,reply ...**ActorMessage) error{
	messageInfo:=&ActorMessageInfo{
		Sender:sender,
		Message:message,
	}
	if len(reply)!=0 {
		messageInfo.reply =reply[0]
	}
	return this.proxy.Emit(this.ID(),messageInfo)
}

