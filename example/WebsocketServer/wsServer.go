package main

import (
	"encoding/json"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
	"reflect"
	"time"
)
//协议对应
var Testid2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}):1,
}

//消息定义
type TestMessage struct {
	Name string
}

//协议接口组
type TestApi struct {
	network.ApiBase
}

func NewTestApi() *TestApi  {
	r:=&TestApi{}
	r.Init(r,nil,Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}

func (this *TestApi)Hello(sess *network.Session,message *TestMessage) error {
	//显示消息内容
	println(fmt.Sprintf("Hello,%s",message.Name))

	//reply
	return this.Reply(sess,&TestMessage{Name:"reply"})
	//或者直接
	//return sess.Emit(1,[]byte("hello client"))
}

func main() {
	s,_:=json.Marshal(&TestMessage{Name:"RockGO"})
	println("直接使用debugClient测试，或者将这条消息复制到下面的websocket在线测试网站测试："+string(s)+
	"\n http://www.blue-zero.com/WebSocket/" +
		"   地址：ws://127.0.0.1:8080/ws")

	conf := &network.ServerConf{
		Proto:                "ws",
		Address:              "0.0.0.0:8080",
		ReadTimeout:          time.Millisecond * 10000,
		OnClientDisconnected: OnDropped,
		OnClientConnected:    OnConnected,
		NetAPI:               NewTestApi(),
	}

	svr := network.NewServer(conf)
	err := svr.Serve()
	if err != nil {
		panic(err)
	}
	<-make(chan struct{})
}

func OnConnected(sess *network.Session) {
	logger.Debug(fmt.Sprintf("client %s connected,session id :%s", sess.RemoteAddr(), sess.ID))
}

func OnDropped(sess *network.Session) {
	logger.Debug(fmt.Sprintf("client %s disconnected,session id :%s", sess.RemoteAddr(), sess.ID))
}