package Actor

import (
	"errors"
	"github.com/zllangct/rockgo/network"
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
	if len(message.Data) == 0 {
		return errors.New("invalid message")
	}
	this.api.Reply(this.session, message.Data[0])
	return nil
}
