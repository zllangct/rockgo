package main

import "time"

type PlayerInfo struct {
	Account string
	Password string
	Name string
	Age  int
	Coin  int64
	LastLoginTime time.Time
}