package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils/gpool"
	"net"
)

type UdpConn struct {
	udpConn   *net.UDPConn
	remoteAddr *net.UDPAddr
}

func (this *UdpConn) WriteMessage(messageType int, data []byte) error  {
	msg := make([]byte, 4)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], uint32(len(msg)))
	if _, err := this.udpConn.WriteToUDP(msg, this.remoteAddr); err != nil {
		logger.Error(fmt.Sprintf("send pkg to %v failed %v", this.remoteAddr, err))
	}
	return nil
}

func (this *UdpConn)Close() error {
	this.udpConn.Close()
	return nil
}

type udpHandler struct {
	conf *ServerConf
	ts   *Server
	conn      *net.UDPConn
	numInvoke int32
	gpool       *gpool.Pool
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
		n, _, err := h.conn.ReadFromUDP(buffer)
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

		handler:= func() {
			ctx := context.Background()
			h.ts.invoke(ctx, pkg[4:])
		}

		if cfg.MaxInvoke > 0 { // use goroutine pool
			if h.gpool == nil {
				h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
			}
			job:=h.gpool.JobPool.Get().(*gpool.Job)
			job.Job = handler
			job.Callback= nil
			h.gpool.JobQueue <-job
		}else {
			go handler()
		}
	}
	return nil
}
