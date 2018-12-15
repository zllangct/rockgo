package main

import (
	"fmt"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
)


/*
协议接口组,需继承network.ApiBase 基类

api结构体中，符合条件的方法会被作为api对客户端服务
api方法的规则为：func 函数名（sess **network.Session, message *消息结构体）
*/

type TestApi struct {
	network.ApiBase
	nodeComponent *Cluster.NodeComponent
	actor  *Actor.ActorComponent
}

//使用协议接口时，需先初始化，初始化时需传入定义的消息号对应字典
//以及所需的消息序列化组件，可轻易切换为protobuf，msgpack等其他序列化工具
func NewTestApi() *TestApi  {
	r:=&TestApi{}
	r.Init(r,Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}



//协议接口 1  Hello
func (this *TestApi)Hello(sess *network.Session,message *TestMessage)  {
	println(fmt.Sprintf("Hello,%s",message.Name))
	p,err:=this.GetParent()
	if err==nil {
		println(fmt.Sprintf("this api parent:%s",p.Name()))
	}

	//reply
	err=sess.Emit(1,[]byte(fmt.Sprintf("hello client %s",message.Name)))
	if err!=nil {
		logger.Error(err)
	}
}

//协议接口 2 创建房间
func (this *TestApi)CreateRoom(sess *network.Session,message *TestCreateRoom)  {
	errReply:= func() {
		r:=&CreateResult{
			Result:false,
		}
		if _,m,err:=this.MessageEncode(r);err ==nil {
			err=sess.Emit(1,m)
		}
	}
	p,err:=this.GetParent()
	if err==nil {
		println(fmt.Sprintf("this api parent:%s",p.Name()))
	}
	actor,err:=this.Actor()
	if err!=nil {
		errReply()
		return
	}
	g,err:=actor.Proxy.GetActorByRole("room")
	if err!=nil {
		errReply()
		return
	}
	roomManager:=Actor.NewActor(g.RndOne(), actor.Proxy)
	var res *Actor.ActorMessage
	err=roomManager.Tell(actor,&Actor.ActorMessage{
		Tittle:"newRoom"},&res)
	if err != nil {
		errReply()
		return
	}
	r:=&CreateResult{
		Result:true,
		RoomID:res.Data[0].(int),
	}
	//reply 创建房间结果反馈到客户端
	if _,m,err:=this.MessageEncode(r);err ==nil {
		err=sess.Emit(1,m)
	}else{
		logger.Error(err)
	}
}

//获取actor组件
func (this *TestApi)Actor() (*Actor.ActorComponent,error) {
	if this.actor==nil{
		p,err:= this.GetParent()
		if err!=nil {
			return nil,err
		}
		err=p.Find(&this.actor)
		if err!=nil {
			return nil,err
		}
	}
	return this.actor,nil
}

//获取node组件
func (this *TestApi)nodeC()(*Cluster.NodeComponent,error){
	if this.nodeComponent == nil{
		o,err:= this.GetParent()
		if err!=nil {
			return nil,err
		}
		err= o.Root().Find(&this.nodeComponent)
		if err!=nil {
			return nil,err
		}
		return this.nodeComponent,nil
	}
	return this.nodeComponent,nil
}