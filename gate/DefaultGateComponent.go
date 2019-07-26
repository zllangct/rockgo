package gate

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/config"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/network"
	"reflect"
	"sync"
	"time"
)

type DefaultGateComponent struct {
	ecs.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	clients       sync.Map // [sessionID,*session]
	NetAPI        network.NetAPI
	server        *network.Server
}

func (this *DefaultGateComponent) IsUnique() int {
	return ecs.UNIQUE_TYPE_GLOBAL
}

func (this *DefaultGateComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}),
	}
	return requires
}

func (this *DefaultGateComponent) Awake(ctx *ecs.Context) {
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}
	if this.NetAPI == nil {
		panic(errors.New("NetAPI is necessity of defaultGateComponent"))
	}
	this.NetAPI.SetParent(this.Parent())
	conf := &network.ServerConf{
		Proto:                "ws",
		PackageProtocol:      &network.TdProtocol{},
		Address:              config.Config.ClusterConfig.NetListenAddress,
		ReadTimeout:          time.Millisecond * time.Duration(config.Config.ClusterConfig.NetConnTimeout),
		OnClientDisconnected: this.OnDropped,
		OnClientConnected:    this.OnConnected,
		NetAPI:               this.NetAPI,
		MaxInvoke:            20,
	}

	this.server = network.NewServer(conf)
	err = this.server.Serve()
	if err != nil {
		panic(err)
	}
}

func (this *DefaultGateComponent) OnConnected(sess *network.Session) {
	this.clients.Store(sess.ID, sess)
	logger.Debug(fmt.Sprintf("client [ %s ] connected,session id :[ %s ]", sess.RemoteAddr(), sess.ID))
}

func (this *DefaultGateComponent) OnDropped(sess *network.Session) {
	this.clients.Delete(sess.ID)
}

func (this *DefaultGateComponent) Destroy() error {
	this.server.Shutdown()
	return nil
}

func (this *DefaultGateComponent) SendMessage(sid string, message interface{}) error {
	if s, ok := this.clients.Load(sid); ok {
		err := this.NetAPI.Reply(s.(*network.Session), message)
		if err != nil {
			return err
		}
	}
	return errors.New(fmt.Sprintf("this session id: [ %s ] not exist", sid))
}

func (this *DefaultGateComponent) Emit(sess *network.Session, message interface{}) error {
	return this.NetAPI.Reply(sess, message)
}
