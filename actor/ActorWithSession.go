package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/rpc"
)

func Upgrade(sess *network.Session,api network.NetAPI,proxy *ActorProxyComponent)(*ActorWithSession ,error ){
	a,ok:=sess.GetProperty("actor")
	if ok {
		return a.(*ActorWithSession), nil
	}
	actor:=&ActorWithSession{
		sess:sess,
		api:api,
	}
	actor.proxy=proxy
	actor.actorID = EmptyActorID()
	err:=proxy.Register(actor)
	if err!=nil {
		return nil, err
	}
	sess.SetProperty("actor",actor)
	sess.AddPostProcessing(func(sess *network.Session) {
		proxy.Unregister(actor)
	})
	return actor, nil
}

type ActorWithSession struct {
	Actor
	sess *network.Session
	api network.NetAPI
}

//调用actor服务
func (this *ActorWithSession)ServiceCall(message *ActorMessage, reply **ActorMessage, role ...string) error  {
	var err error
	service:="service_" + message.Service
	//优先尝试缓存客户端，避免反复查询，尽量去中心化
	g,ok:=this.sess.GetProperty(service)
	if ok {
		err=this.proxy.ServiceCallByRpcClient(this,message,reply,g.(*rpc.TcpClient))
		if err==nil {
			return nil
		}
		this.sess.RemoveProperty(service)
	}
	//无缓存，或者通过缓存调用失败，重新查询调用
   	client,err:= this.proxy.ServiceCallRetrunClient(this,message,reply,role...)
	if err!=nil {
		return err
	}
	if client!=nil {
		this.sess.SetProperty(service,client)
	}
   	return err
}

func (this *ActorWithSession)Tell(sender IActor,message *ActorMessage,reply ...**ActorMessage) error {
	if len(message.Data)!=2 {
		return errors.New("invalid message format")
	}
	return this.api.Reply(this.sess,message)
}