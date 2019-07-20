package Actor

type ActorService struct {
	actor   IActor
	Service string
}

func NewActorService(actor IActor, service string) *ActorService {
	return &ActorService{actor: actor, Service: service}
}

func (this *ActorService) Call(args ...interface{}) ([]interface{}, error) {
	mes := NewActorMessage(this.Service, args...)
	reply := &ActorMessage{}
	err := this.actor.Tell(nil, mes, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}
