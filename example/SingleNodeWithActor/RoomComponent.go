package main

import (
	"errors"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
)

type RoomComponent struct {
	Component.Base
	players       map[int]*Player
	RoomID        int
	actor         Actor.IActor
}

func (this *RoomComponent) Awake() error{
	return nil
}

var ErrLoginPlayerNotExist =errors.New("this player doesnt exist")

func (this *RoomComponent)Enter(player *Player) ([]interface{}, error) {
	this.players[player.UID]=player
	return []interface{}{"hello,welcome to enter this room."},nil
}

func (this *RoomComponent)Actor() (Actor.IActor,error) {
	if this.actor==nil{
		err:= this.Parent.Find(this.actor)
		if err!=nil {
			return nil,err
		}
	}
	return this.actor,nil
}