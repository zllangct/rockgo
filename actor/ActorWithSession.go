package Actor

import (
	"errors"
	"github.com/zllangct/RockGO/network"
)

func Upgrade(sess *network.Session,api network.NetAPI,proxy *ActorProxyComponent)(IActor ,error ){
	a,ok:=sess.GetProperty("actor")
	if ok {
		return a.(IActor), nil
	}
	actor:=&ActorWithSession{
		sess:sess,
		api:api,
	}
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

func (this *ActorWithSession)Tell(sender IActor,message *ActorMessage,reply ...**ActorMessage) error {
	if len(message.Data)!=2 {
		return errors.New("invalid message format")
	}
	return this.api.Reply(this.sess,message)
}