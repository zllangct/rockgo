package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/zllangct/rockgo/network"
	"strconv"
	"sync"
	"time"
)

//MyClient is a example client for tars client testing.
type MyClient struct {
	lock      sync.Mutex
	recvCount int
}

func (c *MyClient) Recv(ctx context.Context, mid uint32, data []byte) {
	c.lock.Lock()
	c.recvCount++
	c.lock.Unlock()
	fmt.Println("recv", string(data))
}

func (s *MyClient) ParseMessage(ctx context.Context, req []byte) ([]uint32, []byte) {
	return []uint32{0}, req
}

//ParsePackage parse package from buff
func (c *MyClient) ParsePackage(buff []byte) (pkgLen, status int) {
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

func getMsg(name string) []byte {
	payload := []byte(name)
	pkg := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(pkg[:4], uint32(len(pkg)))
	copy(pkg[4:], payload)
	return pkg
}

func main() {
	cp := &MyClient{}
	conf := &network.ClientConf{
		Proto:        "tcp",
		ClientProto:  cp,
		QueueLen:     10000,
		IdleTimeout:  time.Second * 5,
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 1000,
		Handler:      cp.Recv,
	}
	client := network.NewClient("localhost:3333", cp, conf)

	count := 500
	go func() {
		name := "Bob"
		for i := 0; i < count; i++ {
			msg := getMsg(name + strconv.Itoa(i))
			println("send:", name+strconv.Itoa(i))
			client.Send(msg)
		}
	}()

	time.Sleep(time.Second * 3)

	cp.lock.Lock()
	println("send count:", count, "recv count:", cp.recvCount)
	cp.lock.Unlock()

	client.Close()
}
