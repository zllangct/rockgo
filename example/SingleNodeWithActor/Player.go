package main

import "github.com/zllangct/RockGO/actor"

type Player struct {
	UID  int         //玩家ID
	Info *PlayerInfo //玩家信息
	addr Actor.IActor
}
