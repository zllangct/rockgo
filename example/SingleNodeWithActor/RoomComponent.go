package main

import (
	"errors"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
)

type RoomComponent struct {
	Component.Base
	Actor.ActorBase
	players map[int]*Player
	RoomID  int
}

func (this *RoomComponent) Initialize() error {
	this.ActorInit(this.Parent(), Actor.ACTOR_TYPE_SYNC)
	return nil
}

func (this *RoomComponent) Awake(ctx *Component.Context) {

}

var ErrLoginPlayerNotExist = errors.New("this player doesnt exist")

func (this *RoomComponent) Enter(player *Player) ([]interface{}, error) {
	this.players[player.UID] = player
	return []interface{}{"hello,welcome to enter this room."}, nil
}
