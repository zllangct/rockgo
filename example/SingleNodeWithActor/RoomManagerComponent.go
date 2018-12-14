package main

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
)

type RoomManagerComponent struct {
	Component.Base
	rooms map[int]*RoomComponent
	messageHandler map[string]func(message *Actor.ActorMessageInfo)
	increasing int   //实际运用不这样
}

func (this *RoomManagerComponent)Awake()error{
	this.rooms = make(map[int]*RoomComponent)
	this.messageHandler=map[string]func(message *Actor.ActorMessageInfo){
		"newRoom":this.NewRoom,
		"enter":this.EnterRoom,
	}
	return nil
}

func (this *RoomManagerComponent) MessageHandlers() map[string]func(message *Actor.ActorMessageInfo) {
	return this.messageHandler
}

func (this *RoomManagerComponent)NewRoom(message *Actor.ActorMessageInfo)  {
	c:=&RoomComponent{}
	this.Parent.AddComponent(c)
	this.increasing++
	c.RoomID=this.increasing
	this.rooms[c.RoomID]=c
	message.Reply=&Actor.ActorMessage{
		Data:[]interface{}{c.RoomID},
	}
}

func (this *RoomManagerComponent)EnterRoom(message *Actor.ActorMessageInfo)  {
	roomID:=message.Message.Data[0].(int)
	UID:=message.Message.Data[1].(int)
	player:=&Player{UID:UID}
	if room,ok:=this.rooms[roomID];ok{
		err:=room.Enter(player)
		if err!=nil {
			*message.Reply=Actor.ActorMessage{
				Data:[]interface{}{false},
			}
		}
	}else{
		*message.Reply=Actor.ActorMessage{
			Data:[]interface{}{false},
		}
	}
	*message.Reply=Actor.ActorMessage{
		Data:[]interface{}{true},
	}
}