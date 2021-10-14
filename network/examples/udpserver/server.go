package main

import (
	"context"
	"fmt"
	"github.com/zllangct/rockgo/network"
	"time"
)

//MyServer testing tars udp server
type MyServer struct{}

//ParseMessage recv package and make response.
func (s *MyServer) ParseMessage(ctx context.Context, req []byte) ([]uint32, []byte) {
	return []uint32{0}, req
}

//ParsePackage parse full tars package.
func (s *MyServer) ParsePackage(buff []byte) (pkgLen, status int) {
	return len(buff), network.PACKAGE_FULL
}

func (s *MyServer) Recv(sess *network.Session, data []byte) {
	fmt.Println("recv", string(data))
	sess.Emit(0, []byte("yep  "+string(data)))
}

func main() {
	s := MyServer{}
	conf := &network.ServerConf{
		Proto:           "udp",
		PackageProtocol: &s,
		Address:         "127.0.0.1:3333",
		//MaxAccept:     500,
		PoolMode:      true,
		MaxInvoke:     20,
		Handler:       s.Recv,
		AcceptTimeout: time.Millisecond * 500,
		ReadTimeout:   time.Millisecond * 100,
		WriteTimeout:  time.Millisecond * 100,
		IdleTimeout:   time.Millisecond * 600000,
	}

	svr := network.NewServer(conf)
	err := svr.Serve()
	if err != nil {
		fmt.Println(err)
	}
}
