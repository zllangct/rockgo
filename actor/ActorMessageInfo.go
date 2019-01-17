package Actor

import "errors"

/*
	ActorMessage Information
*/

var ErrTimeout =errors.New("actor tell time out")

type ActorMessageInfo struct {
	Sender  IActor
	Message *ActorMessage
	reply   **ActorMessage
	done    chan struct{}
	isDone  bool
	err     error
}

func (this *ActorMessageInfo)NeedReply(isReply bool)  {
	if isReply {
		this.done = make(chan struct{})
	}
}

func (this *ActorMessageInfo)IsNeedReply() bool {
	return this.done != nil
}

func (this *ActorMessageInfo)Reply(tittle string,args ...interface{})  {
	if this.done!=nil {
		*this.reply=&ActorMessage{
			Service: tittle,
			Data:    args,
		}
		this.done<- struct{}{}
		this.isDone =true
	}
}

func (this *ActorMessageInfo) ReplyError(err error)  {
	this.err=err
	if this.done!=nil {
		this.done<- struct{}{}
		this.isDone =true
	}
}

func (this *ActorMessageInfo) ReplyVoid()  {
	if this.done!=nil && !this.isDone{
		this.done<- struct{}{}
		this.isDone=true
	}
}

type ActorRpcMessageInfo struct {
	Target  ActorID
	Sender ActorID
	Message *ActorMessage
}

