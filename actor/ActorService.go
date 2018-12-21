package Actor

type ActorProxyService struct {
	proxy *ActorProxyComponent
}

func (this *ActorProxyService) init(proxy *ActorProxyComponent) {
	this.proxy = proxy
}
func (this *ActorProxyService) Tell(args *ActorRpcMessageInfo, reply *ActorMessage) error {
	minfo := &ActorMessageInfo{
		Sender:  &Actor{proxy: this.proxy, actorID: args.Sender},
		Message: args.Message,
		reply:   &reply,
	}
	return this.proxy.Emit(args.Target, minfo)
}

type ServiceCall struct {
	Sender  ActorID
	Message *ActorMessage
}

func (this *ActorProxyService) ServiceCall(args *ServiceCall, reply *ActorMessage) error {
	sender:=&Actor{proxy: this.proxy, actorID: args.Sender}
	var r *ActorMessage
	err:= this.proxy.ServiceCall(sender,args.Message,&r)
	if err!=nil{
		return err
	}
	*reply = *r
	return nil
}
