package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/zllangct/rockgo/logger"
	"github.com/zllangct/rockgo/utils/UUID"
	"io"
	"net"
	"reflect"
	"sync/atomic"
	"time"
)

type TcpConn struct {
	tcpConn *net.TCPConn
}

func (this *TcpConn) Addr() string {
	return this.tcpConn.RemoteAddr().String()
}

func (this *TcpConn) WriteMessage(messageType uint32, data []byte) error {
	msg := make([]byte, 8)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], uint32(len(msg)))
	binary.BigEndian.PutUint32(msg[4:8], messageType)
	if _, err := this.tcpConn.Write(msg); err != nil {
		logger.Error(fmt.Sprintf("send pkg to %v failed %v", this.tcpConn.RemoteAddr(), err))
	}
	return nil
}

func (this *TcpConn) Close() error {
	return this.tcpConn.Close()
}

type tcpHandler struct {
	conf *ServerConf

	lis         *net.TCPListener
	ts          *Server
	acceptNum   int32
	invokeNum   int32
	readBuffer  int
	writeBuffer int
	tcpNoDelay  bool
	idleTime    time.Time
	gpool       *Pool
}

func (h *tcpHandler) Listen() (err error) {
	cfg := h.conf
	addr, err := net.ResolveTCPAddr("tcp4", cfg.Address)
	if err != nil {
		return err
	}
	h.lis, err = net.ListenTCP("tcp4", addr)
	logger.Info(fmt.Sprintf("TCP server listening and serving TCP on: [ %s ]", cfg.Address))
	return
}

func (h *tcpHandler) Handle() error {
	conf := h.conf
	//对象池模式下，初始pool大小为20
	if conf.PoolMode && conf.MaxInvoke == 0 {
		conf.MaxInvoke = 20
	}
	h.gpool = GetGlobalPool(int(conf.MaxInvoke), conf.QueueCap)

	for !h.ts.isClosed {
		if conf.AcceptTimeout != 0 {
			h.lis.SetDeadline(time.Now().Add(conf.AcceptTimeout)) // set accept timeout
		}
		conn, err := h.lis.AcceptTCP()
		if err != nil {
			if !isNoDataError(err) {
				logger.Error(fmt.Sprintf("Accept error: %v", err))
			} else if conn != nil {
				conn.SetKeepAlive(true)
			}
			continue
		}
		go func(conn *net.TCPConn) {
			logger.Debug("TCP accept:", conn.RemoteAddr())
			atomic.AddInt32(&h.acceptNum, 1)
			_ = conn.SetReadBuffer(conf.TCPReadBuffer)
			_ = conn.SetWriteBuffer(conf.TCPWriteBuffer)
			_ = conn.SetNoDelay(conf.TCPNoDelay)
			sess := &Session{
				ID:         UUID.Next(),
				properties: make(map[string]interface{}),
				conn:       &TcpConn{tcpConn: conn},
			}
			if h.conf.OnClientConnected != nil {
				//TODO 异常处理
				h.conf.OnClientConnected(sess)
			}
			h.recv(sess, conn)
			sess.locker.Lock()
			sess.conn = nil
			sess.locker.Unlock()
			if h.conf.OnClientDisconnected != nil {
				h.conf.OnClientDisconnected(sess)
			}
			atomic.AddInt32(&h.acceptNum, -1)
		}(conn)
	}
	if h.gpool != nil {
		h.gpool.Release()
	}
	return nil
}

func (h *tcpHandler) recv(sess *Session, conn *net.TCPConn) {
	defer conn.Close()

	sess.SetProperty("workerID", WORKER_ID_RANDOM)

	cfg := h.conf
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	h.idleTime = time.Now()
	var n int
	var err error
	//TODO 添加TCP频率控制
	for !h.ts.isClosed {
		if cfg.ReadTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		}
		n, err = conn.Read(buffer)
		if err != nil {
			if len(currBuffer) == 0 && h.ts.numInvoke == 0 && h.idleTime.Add(cfg.IdleTimeout).Before(time.Now()) {
				return
			}
			h.idleTime = time.Now()
			if isNoDataError(err) {
				continue
			}
			if err == io.EOF {
				logger.Debug("connection closed by remote:", conn.RemoteAddr())
			} else {
				logger.Error("read package error:", reflect.TypeOf(err), err)
			}
			return
		}

		currBuffer = append(currBuffer, buffer[:n]...)
		for {
			pkgLen, status := h.ts.conf.PackageProtocol.ParsePackage(currBuffer)
			if status == PACKAGE_LESS {
				break
			}
			if status == PACKAGE_FULL {
				pkg := make([]byte, pkgLen-4)
				copy(pkg, currBuffer[4:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				// use goroutine pool
				if h.conf.PoolMode {
					var wid int32
					var ok bool
					m,PropertyOk:=sess.GetProperty("workerID")
					if wid,ok=m.(int32);!PropertyOk || !ok{
						wid = WORKER_ID_RANDOM
					}
					h.gpool.AddJobFixed(h.handler, []interface{}{sess, pkg}, wid)
				} else {
					go h.handler(nil,sess, pkg)
				}
				if len(currBuffer) > 0 {
					continue
				}
				currBuffer = nil
				break
			}
			logger.Error(fmt.Sprintf("parse package error %s %v", conn.RemoteAddr(), err))
			return
		}
	}
}

func (h *tcpHandler) handler(poolCtx []interface{},args ...interface{}) {
	if poolCtx != nil && len(poolCtx)>0 {
		args[0].(*Session).SetProperty("workerID", poolCtx[0].(int32))
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cid", args[0])
	if h.conf.Handler != nil {
		h.conf.Handler(args[0].(*Session), args[1].([]byte))
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
