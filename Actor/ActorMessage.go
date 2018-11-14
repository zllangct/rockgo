package Actor

import (

)

type ActorMessageBase struct {
	title string

}

func (this *ActorMessageBase)GetTitle() string {
	return this.title
}
