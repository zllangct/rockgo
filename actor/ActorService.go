package Actor

type ActorProxyService struct {
	proxy *ActorProxyComponent
}

func (this *ActorProxyService)init(proxy *ActorProxyComponent)  {
	this.proxy= proxy
}
func (this *ActorProxyService)Tell(args *ActorRpcMessageInfo,reply *ActorMessage) error {
	minfo:=&ActorMessageInfo{
		Sender:  &Actor{proxy: this.proxy,actorID:args.Sender},
		Message: args.Message,
		reply:   &reply,
	}
	return this.proxy.Emit(args.Target,minfo)
}
