package Actor

import "github.com/zllangct/RockGO/Network"

/*
	ActorMessage Information
 */

type ActorMessageInfo struct {
	Sender     Actor
	Recipients Actor
	Session    *Network.Session
	Message    IActorMessage
}

func (this *ActorMessageInfo)Reply(message IActorMessage){
	this.Recipients.Emit(&ActorMessageInfo{
		Sender:     this.Recipients,
		Recipients: this.Sender,
		Session:    this.Session,
		Message:    message,
	})
}