package Actor

type ActorMessage struct {
	Service string
	Data    []interface{}
}

func NewActorMessage(service string, args ...interface{}) *ActorMessage {
	return &ActorMessage{
		Service: service,
		Data:    args,
	}
}
