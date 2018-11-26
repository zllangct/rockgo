package Actor

import "sync"

type ActorProxy struct {
	actors sync.Map //[ActorID,*Actor]
}

func (this *ActorProxy)Register(actor *ActorComponent)  {
	if _,ok:=this.actors.LoadOrStore(actor.ActorID,actor) ;ok{
		return
	}
	//如果是远程Actor,则注册ActorID到分布式  //TODO 需要在actor销毁时在分布式上取消注册
	if actor.ActorType == ACTOR_TYPE_REMOTE {

	}
}

func (this *ActorProxy)Unregister(actor *ActorComponent)  {

}