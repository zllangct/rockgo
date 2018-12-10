package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils/UUID"
	"github.com/zllangct/RockGO/utils/current"
	"github.com/zllangct/RockGO/utils/gpool"
	"io"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"time"


)

type TcpConn struct {
	tcpConn   *net.TCPConn
}

func (this *TcpConn)Addr()string  {
	return this.tcpConn.RemoteAddr().String()
}

func (this *TcpConn) WriteMessage(messageType uint32, data []byte) error  {
	msg := make([]byte, 8)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], uint32(len(msg)))
	binary.BigEndian.PutUint32(msg[4:8], messageType)
	if _, err := this.tcpConn.Write(msg); err != nil {
		logger.Error(fmt.Sprintf("send pkg to %v failed %v", this.tcpConn.RemoteAddr(), err))
	}
	return nil
}

func (this *TcpConn)Close() error {
	this.tcpConn.Close()
	return nil
}

type tcpHandler struct {
	conf *ServerConf

	lis *net.TCPListener
	ts  *Server
	acceptNum   int32
	invokeNum   int32
	readBuffer  int
	writeBuffer int
	tcpNoDelay  bool
	idleTime    time.Time
	gpool       *gpool.Pool
}

func (h *tcpHandler) Listen() (err error) {
	cfg := h.conf
	addr, err := net.ResolveTCPAddr("tcp4", cfg.Address)
	if err != nil {
		return err
	}
	h.lis, err = net.ListenTCP("tcp4", addr)
	logger.Info("Listening on", cfg.Address)
	return
}

func (h *tcpHandler) handleConn(sess *Session, conn *net.TCPConn, pkg []byte) {
	handler := func() {
		remoteAddr := conn.RemoteAddr().String()
		ipPort := strings.Split(remoteAddr, ":")
		ctx := current.ContextWithCurrent(context.Background())
		ok := current.SetClientIPWithContext(ctx, ipPort[0])
		ctx =context.WithValue(ctx,"sess",sess)
		if !ok {
			logger.Error("Failed to set context with client ip")
		}
		ok = current.SetClientPortWithContext(ctx, ipPort[1])
		if !ok {
			logger.Error("Failed to set context with client port")
		}
		mid,data:= h.conf.PackageProtocol.ParseMessage(ctx, pkg)
		h.ts.invoke(ctx, mid[0],data)
	}

	workerID,ok:=sess.GetProperty("workerID")
	if !ok {
		workerID = -1
	}

	cfg := h.conf
	if cfg.MaxInvoke > 0 { // use goroutine pool
		if h.gpool == nil {
			h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
		}
		job:=h.gpool.JobPool.Get().(*gpool.Job)
		job.WorkerID = workerID.(int32)
		job.Job = handler
		job.Callback= func(w int32){
			sess.SetProperty("workerID",w)
		}
		h.gpool.JobQueue <-job
	} else {
		go handler()
	}
}

func (h *tcpHandler) Handle() error {
	cfg := h.conf
	for !h.ts.isClosed {
		if cfg.AcceptTimeout !=0 {
			h.lis.SetDeadline(time.Now().Add(cfg.AcceptTimeout)) // set accept timeout
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
			_=conn.SetReadBuffer(cfg.TCPReadBuffer)
			_=conn.SetWriteBuffer(cfg.TCPWriteBuffer)
			_=conn.SetNoDelay(cfg.TCPNoDelay)
			sess:=&Session{
				ID:UUID.Next(),
				properties:make( map[string]interface{}),
				conn:&TcpConn{tcpConn:conn},
			}
			if h.conf.OnClientConnected!=nil {
				h.conf.OnClientConnected(sess)
			}
			h.recv(sess, conn)
			if h.conf.OnClientDisconnected!=nil{
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
	sess.SetProperty("workerID",-1)

	cfg := h.conf
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	h.idleTime = time.Now()
	var n int
	var err error
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
				h.handleConn(sess,conn, pkg)
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
