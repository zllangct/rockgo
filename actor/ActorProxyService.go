package Actor

type ActorProxyService struct {
	proxy *ActorProxyComponent
}

func (this *ActorProxyService) init(proxy *ActorProxyComponent) {
	this.proxy = proxy
}
func (this *ActorProxyService) Tell(args *ActorRpcMessageInfo, reply *ActorMessage) error {
	minfo := &ActorMessageInfo{
		Sender: NewActor(args.Sender,this.proxy),
		Message: args.Message,
		reply:   &reply,
	}
	return this.proxy.Emit(args.Target, minfo)
}

func (this *ActorProxyService) ServiceInquiry(service string, reply *ActorID) error {
	r,err:= this.proxy.GetLocalActorService(service)
	if err!=nil {
		return err
	}
	*reply = r.actor.ID()
	return nil
}