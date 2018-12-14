package Actor

type ActorProxyService struct {
	proxy *ActorProxyComponent
}

func (this *ActorProxyService)init(proxy *ActorProxyComponent)  {
	this.proxy= proxy
}
func (this *ActorProxyService)Tell(args *ActorRpcMessageInfo,reply *ActorMessage) error {
	return this.proxy.Emit(args.Target,&ActorMessageInfo{
		Sender:&ActorRemote{proxy:this.proxy,actorID:args.Sender},
		Message: args.Message,
		Reply:reply,
	})
}
