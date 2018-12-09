package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/zllangct/RockGO/network"
	"time"

)

//MyServer testing tars udp server
type MyServer struct{}

//Invoke recv package and make response.
func (s *MyServer) Invoke(ctx context.Context,req []byte) {
	println(string(req))
}

//ParsePackage parse full tars package.
func (s *MyServer) ParsePackage(buff []byte) (pkgLen, status int) {
	if len(buff) < 4 {
		return 0, network.PACKAGE_LESS
	}
	length := binary.BigEndian.Uint32(buff[:4])

	if length > 1048576000 || len(buff) > 1048576000 { // 1000MB
		return 0, network.PACKAGE_ERROR
	}
	if len(buff) < int(length) {
		return 0, network.PACKAGE_LESS
	}
	return int(length), network.PACKAGE_FULL
}
func main() {
	conf := &network.ServerConf{
		Proto:   "udp",
		Address: "127.0.0.1:3333",
		//MaxAccept:     500,
		MaxInvoke:     20,
		AcceptTimeout: time.Millisecond * 500,
		ReadTimeout:   time.Millisecond * 100,
		WriteTimeout:  time.Millisecond * 100,
		IdleTimeout:   time.Millisecond * 600000,
	}
	s := MyServer{}
	svr := network.NewServer(&s, conf)
	err := svr.Serve()
	if err != nil {
		fmt.Println(err)
	}
}
