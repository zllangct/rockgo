package Actor

type ActorService struct {
	proxy *ActorProxyComponent
}

func (this *ActorService)init(proxy *ActorProxyComponent)  {
	this.proxy= proxy
}
func (this *ActorService)Tell(args *ActorRpcMessageInfo,reply *ActorMessage) error {
	return this.proxy.Emit(args.Target,&ActorMessageInfo{
		Sender:&ActorRemote{proxy:this.proxy,actorID:args.Sender},
		Message: args.Message,
	},reply)
}