package gate

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/network"
	"reflect"
	"sync"
	"time"
)

type GateComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	clients       sync.Map // [sessionID,*session]
	NetAPI        network.NetAPI
}

func (this *GateComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *GateComponent) Awake() {
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Fatal("get node component failed")
		panic(err)
		return
	}
	conf := &network.ServerConf{
		Proto:                "tcp",
		Address:              Config.Config.ClusterConfig.NetListenAddress,
		ReadTimeout:          time.Millisecond * time.Duration(Config.Config.ClusterConfig.NetConnTimeout),
		OnClientDisconnected: this.OnDropped,
		OnClientConnected:    this.OnConnected,
		NetAPI:               this.NetAPI,
	}

	svr := network.NewServer(conf)
	err = svr.Serve()
	if err != nil {
		panic(err)
	}
}

func (this *GateComponent) OnConnected(sess *network.Session) {
	this.clients.Store(sess.ID, sess)
	logger.Debug(fmt.Sprintf("client %s connected,session id :%s", sess.RemoteAddr(), sess.ID))
}

func (this *GateComponent) OnDropped(sess *network.Session) {
	this.clients.Delete(sess.ID)
}

func (this *GateComponent) SendMessage(sid string, message interface{}) error {
	if s, ok := this.clients.Load(sid); ok {
		mid, b, err := this.NetAPI.MessageEncode(message)
		if err != nil {
			return err
		}
		s.(*network.Session).Emit(mid, b)
	}
	return errors.New(fmt.Sprintf("this session id: %s not exist", sid))
}

func (this *GateComponent) Emit(sess *network.Session, message interface{}) error {
	mid, b, err := this.NetAPI.MessageEncode(message)
	if err != nil {
		return err
	}
	sess.Emit(mid, b)
	return nil
}
