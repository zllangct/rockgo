package Actor

/*
	ActorMessage Information
*/

type ActorMessageInfo struct {
	Sender  IActor
	Message *ActorMessage
	reply   **ActorMessage
	done    chan struct{}
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
	}
}

func (this *ActorMessageInfo)CallError(err error)  {
	this.err=err
	if this.done!=nil {
		this.done<- struct{}{}
	}
}

type ActorRpcMessageInfo struct {
	Target  ActorID
	Sender ActorID
	Message *ActorMessage
	//Reply *ActorMessage
}

