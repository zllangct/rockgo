package Actor

type ActorBase struct {
	messageHandler map[string]func(message *ActorMessageInfo)
}

func (this *ActorBase) MessageHandlers() map[string]func(message *ActorMessageInfo) {
	return this.messageHandler
}