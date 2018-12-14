package Actor



type IActor interface {
	Tell(sender IActor,message *ActorMessage,reply ...*ActorMessage) error
	ID() ActorID
}

type IActorMessageHandler interface {
	MessageHandlers()map[string]func(message *ActorMessageInfo)
}

type ActorRemote struct {
	actorID ActorID
	proxy *ActorProxyComponent
}

func NewActor(id ActorID,proxy *ActorProxyComponent) IActor {
	return &ActorRemote{
		actorID:id,
		proxy:proxy,
	}
}

func (this *ActorRemote)ID ()ActorID{
	return this.actorID
}

func (this *ActorRemote)Tell(sender IActor,message *ActorMessage,reply ...*ActorMessage) error{
	messageInfo:=&ActorMessageInfo{
		Sender:sender,
		Message:message,
	}
	if len(reply)==0 {
		messageInfo.Reply= new(ActorMessage)
	}else{
		messageInfo.Reply=reply[0]
	}
	return this.proxy.Emit(this.ID(),messageInfo)
}