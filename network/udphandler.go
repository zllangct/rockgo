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
	"sync"
	"sync/atomic"
	"time"
)

var ErrUdpConnClosed = errors.New("this udp conn is closed")

type UdpConn struct {
	udpConn       *net.UDPConn
	remoteAddr    *net.UDPAddr
	cid           uint32
	timeout       <-chan struct{}
	closeCallback func()
	m             *sync.Map
	once          *sync.Once
}

func (this *UdpConn) Addr() string {
	return this.remoteAddr.String()
}

func (this *UdpConn) Init() {
	go func() {
		<-this.timeout
		this.Close()
		this.m.Delete(this.cid)
	}()
}

func (this *UdpConn) SetReadDeadline(duration time.Duration) {
	this.once.Do(this.Init)
	this.timeout = timer.After(duration)
}

func (this *UdpConn) WriteMessage(messageType uint32, data []byte) error {
	msg := make([]byte, 12)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], uint32(len(msg)))
	binary.BigEndian.PutUint32(msg[4:8], this.cid)
	binary.BigEndian.PutUint32(msg[8:12], messageType)
	if _, err := this.udpConn.WriteToUDP(msg, this.remoteAddr); err != nil {
		logger.Error(fmt.Sprintf("send pkg to %v failed %v", this.remoteAddr, err))
	}
	return nil
}

func (this *UdpConn) Close() error {
	this.remoteAddr = nil
	return nil
}

type udpHandler struct {
	conf      *ServerConf
	ts        *Server
	conn      *net.UDPConn
	conns     *sync.Map
	numInvoke int32
	cid       uint32
	gpool     *gpool.Pool
}

func (h *udpHandler) Listen() error {
	conf := h.conf
	//对象池模式下，初始pool大小为20
	if conf.PoolMode && conf.MaxInvoke == 0 {
		conf.MaxInvoke = 20
	}
	h.gpool = gpool.GetGloblePool(int(conf.MaxInvoke), conf.QueueCap)

	addr, err := net.ResolveUDPAddr("udp", conf.Address)
	if err != nil {
		return err
	}
	h.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("UDP server listening and serving UDP on: [ %s ]", h.conn.LocalAddr()))
	return nil
}

func (h *udpHandler) Handle() error {

	wg := sync.WaitGroup{}
	buffer := make([]byte, 65535)
	for {
		wg.Wait()
		if h.ts.isClosed {
			return nil
		}
		wg.Add(1)
		go func() {
			n, udpAddr, err := h.conn.ReadFromUDP(buffer)
			if err != nil {
				if !isNoDataError(err) {
					logger.Error(fmt.Sprintf("Close connection %s: %v", h.conf.Address, err))
					return
				}
			}
			data := make([]byte, n)
			copy(data, buffer[0:n])
			wg.Done()

			if h.conf.Handler != nil {
				h.conf.Handler( &Session{
					conn:       &UdpConn{remoteAddr: udpAddr, udpConn: h.conn,m:h.conns},
				}, data)
				return
			}

			var new bool
			cfg := h.conf
			ctx := context.Background()
			mid, pkg := h.conf.PackageProtocol.ParseMessage(ctx, data)

			if len(mid) != 2 {
				logger.Warn("udp data fmt incorrect")
				return
			}

			cid := mid[0]
			if cid == 0 {
				cid = atomic.AddUint32(&h.cid, 1)
				new = true
			}

			s, _ := h.conns.LoadOrStore(cid, &Session{
				ID:         UUID.Next(),
				properties: make(map[string]interface{}),
				conn:       &UdpConn{remoteAddr: udpAddr, udpConn: h.conn,m:h.conns},
			})
			sess := s.(*Session)
			sess.conn.(*UdpConn).SetReadDeadline(cfg.ReadTimeout)

			wid := int32(-1)
			item, ok := sess.GetProperty("workerID")
			if ok {
				wid = item.(int32)
				sess.SetProperty("workerID", int32(-1))
			}

			if new {
				//异常处理
				h.ts.conf.OnClientConnected(sess)
				sess.conn.(*UdpConn).closeCallback = func() {
					h.ts.conf.OnClientDisconnected(sess)
				}
			}

			if h.conf.NetAPI != nil && mid != nil {
				// use goroutine pool
				if h.conf.PoolMode {
					h.gpool.AddJobSerial(h.handler, []interface{}{sess, pkg}, wid, func(workerID int32) {
						wid = workerID
					})
				} else {
					go h.handler(sess, pkg)
				}
			} else {
				logger.Error("no message handler")
				return
			}
		}()

	}
}

func (h *udpHandler) handler(args ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cid", args[0])
	if h.conf.Handler != nil {
		h.conf.Handler(args[1].(*Session), args[1].([]byte))
	} else {
		mid, mes := h.conf.PackageProtocol.ParseMessage(ctx, args[1].([]byte))
		if h.conf.NetAPI != nil && mid != nil {
			h.ts.invoke(ctx, mid[0], mes)
		} else {
			logger.Error("no message handler")
			return
		}
	}
}
