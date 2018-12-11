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

type DefaultGateComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	clients       sync.Map // [sessionID,*session]
	NetAPI        network.NetAPI
}

func (this *DefaultGateComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *DefaultGateComponent) Awake() {
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Fatal("get node component failed")
		panic(err)
		return
	}
	if this.NetAPI==nil {
		panic(errors.New("NetAPI is necessity of defaultGateComponent"))
	}
	this.NetAPI.SetParent(this.Parent)
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

func (this *DefaultGateComponent) OnConnected(sess *network.Session) {
	this.clients.Store(sess.ID, sess)
	logger.Debug(fmt.Sprintf("client %s connected,session id :%s", sess.RemoteAddr(), sess.ID))
}

func (this *DefaultGateComponent) OnDropped(sess *network.Session) {
	this.clients.Delete(sess.ID)
}

func (this *DefaultGateComponent) SendMessage(sid string, message interface{}) error {
	if s, ok := this.clients.Load(sid); ok {
		mid, b, err := this.NetAPI.MessageEncode(message)
		if err != nil {
			return err
		}
		s.(*network.Session).Emit(mid, b)
	}
	return errors.New(fmt.Sprintf("this session id: %s not exist", sid))
}

func (this *DefaultGateComponent) Emit(sess *network.Session, message interface{}) error {
	mid, b, err := this.NetAPI.MessageEncode(message)
	if err != nil {
		return err
	}
	sess.Emit(mid, b)
	return nil
}
