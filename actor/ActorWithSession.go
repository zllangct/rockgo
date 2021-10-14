package Actor

import (
	"github.com/zllangct/RockGO/network"
	"sync"
)


type ActorWithSession struct {
	locker   sync.RWMutex
	actorID 	     ActorID
	proxy    *ActorProxyComponent
	session  *network.Session
	api      network.NetAPI
}

func NewActorWithSession(proxy *ActorProxyComponent, sess *network.Session) (*ActorWithSession, error) {
	actor := &ActorWithSession{ actorID: EmptyActorID(), proxy: proxy, session: sess}
	err := proxy.Register(actor)
	if err != nil {
		return nil, err
	}
	actor.session.AddPostProcessing(func(sess *network.Session) {
		proxy.Unregister(actor)
	})

	return actor, nil
}

func (this *ActorWithSession) ID() ActorID {
	return this.actorID
}

func (this *ActorWithSession) Tell(sender IActor, message *ActorMessage, reply ...**ActorMessage) error{
	this.api.Reply(this.session, message)
	return nil
}
