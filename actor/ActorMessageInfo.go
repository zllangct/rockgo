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

func (this *ActorMessageInfo) Reply(args ...interface{}) error  {
	return this.ReplyWithTittle("",args...)
}

func (this *ActorMessageInfo) ReplyWithTittle(tittle string,args ...interface{}) error  {
	if this.done!=nil {
		*this.reply=&ActorMessage{
			Service: tittle,
			Data:    args,
		}
		return nil
	}else{
		return errors.New("this message invalid")
	}
}

func (this *ActorMessageInfo) replyError(err error)  {
	this.err=err
	if this.done!=nil && !this.isDone{
		this.done<- struct{}{}
		this.isDone =true
	}
}

func (this *ActorMessageInfo) replySuccess()  {
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

