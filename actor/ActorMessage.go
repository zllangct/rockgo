package Actor

type IActorMessage interface {
	Tittle()string
	Data()interface{}
}


type ActorMessage struct {
	title string
	data interface{}
}

func (this *ActorMessage)Tittle() string {
	return this.title
}

func (this *ActorMessage)Data() interface{} {
	return this.data
}