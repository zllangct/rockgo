package main

import "github.com/zllangct/RockGO"

func main()  {
	server := RockGO.DefaultNode()
	server.Serve()
}
