package network

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/timer"
	"github.com/zllangct/RockGO/utils/UUID"
	"github.com/zllangct/RockGO/utils/gpool"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

var ErrUdpConnClosed =errors.New("this udp conn is closed")

type UdpConn struct {
	udpConn   *net.UDPConn
	remoteAddr *net.UDPAddr
	sess       uint32
	lock	   sync.Mutex
	timeout    <-chan struct{}
	closeCallback func()
	m			*sync.Map
}
func (this *UdpConn)Addr()string  {
	return this.remoteAddr.String()
}

func (this *UdpConn)Init(){
	go func() {
		<-this.timeout
		this.Close()
		this.m.Delete(this.sess)
	}()
}

func (this *UdpConn)SetReadDeadline(duration time.Duration)  {
	this.timeout=timer.After(duration)
}

func (this *UdpConn) WriteMessage(messageType uint32, data []byte) error  {
	msg := make([]byte, 12)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], uint32(len(msg)))
	binary.BigEndian.PutUint32(msg[4:8], this.sess)
	binary.BigEndian.PutUint32(msg[8:12], messageType)
	if _, err := this.udpConn.WriteToUDP(msg, this.remoteAddr); err != nil {
		logger.Error(fmt.Sprintf("send pkg to %v failed %v", this.remoteAddr, err))
	}
	return nil
}

func (this *UdpConn)Close() error {
	this.remoteAddr=nil
	return nil
}

type udpHandler struct {
	conf *ServerConf
	ts   *Server
	conn      *net.UDPConn
	conns      *sync.Map
	numInvoke int32
	cid        uint32
	gpool      *gpool.Pool
}

func (h *udpHandler) Listen() error {
	cfg := h.conf
	addr, err := net.ResolveUDPAddr("udp4", cfg.Address)
	if err != nil {
		return err
	}
	h.conn, err = net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	logger.Info("UDP listen", h.conn.LocalAddr())
	return nil
}

func (h *udpHandler) Handle() error {
	buffer := make([]byte, 65535)
	for !h.ts.isClosed {
		n, udpAddr, err := h.conn.ReadFromUDP(buffer)
		if err != nil {
			if isNoDataError(err) {
				continue
			} else {
				logger.Error(fmt.Sprintf("Close connection %s: %v", h.conf.Address, err))
				return err
			}
		}
		pkg := make([]byte, n)
		copy(pkg, buffer[0:n])
		cfg := h.conf
		ctx := context.Background()
		mid,data:= h.conf.PackageProtocol.ParseMessage(ctx, pkg)

		handler:= func() {
			if len(mid)!=2 {
				h.ts.Shutdown()
				panic(errors.New(fmt.Sprintf("this package protoc %s doesnt fit",reflect.TypeOf(h.conf.PackageProtocol).Name())))
			}
			var new bool
			cid :=mid[0]
			if cid ==0 {
				cid = atomic.AddUint32(&h.cid,1)
				new =true
			}

			s,_:=h.conns.LoadOrStore(cid,&Session{
				ID:UUID.Next(),
				properties:make( map[string]interface{}),
				conn:&UdpConn{remoteAddr:udpAddr,udpConn:h.conn},
			})
			sess:=s.(*Session)
			sess.conn.(*UdpConn).SetReadDeadline(cfg.ReadTimeout)

			if new {
				h.ts.conf.OnClientConnected(sess)
				sess.conn.(*UdpConn).closeCallback= func() {
					h.ts.conf.OnClientDisconnected(sess)
				}
			}

			ctx=context.WithValue(ctx,"sess",sess)
			h.ts.invoke(ctx,mid[1],data)
		}

		if cfg.MaxInvoke > 0 { // use goroutine pool
			go func() {
				if h.gpool == nil {
					h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
				}
				job:=h.gpool.JobPool.Get().(*gpool.Job)
				job.Job = handler
				job.Callback= nil
				h.gpool.JobQueue <-job
			}()
		}else {
			go handler()
		}
	}
	return nil
}
