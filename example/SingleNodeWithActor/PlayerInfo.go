package main

import "time"

type PlayerInfo struct {
	UID int
	Account string
	Password string
	Name string
	Age  int
	Coin  int64
	LastLoginTime time.Time
}