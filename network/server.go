package network

import (
	"context"
	"sync/atomic"
	"time"
)

const (
	//PACKAGE_LESS shows is not a completed package.
	PACKAGE_LESS = iota
	//PACKAGE_FULL shows is a completed package.
	PACKAGE_FULL
	//PACKAGE_ERROR shows is a error package.
	PACKAGE_ERROR
)
//Protocol is interface for handling the server side tars package.
type Protocol interface {
	Invoke(ctx context.Context, pkg []byte)
	ParsePackage(buff []byte) (int, int)
}

//ServerHandler  is interface with listen and handler method
type ServerHandler interface {
	Listen() error
	Handle() error
}

//ServerConf server config for tars server side.
type ServerConf struct {
	Proto          string
	Address        string
	MaxInvoke      int32
	AcceptTimeout  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	QueueCap       int
	TCPReadBuffer  int
	TCPWriteBuffer int
	TCPNoDelay     bool
}

//Server tars server struct.
type Server struct {
	svr        Protocol
	conf       *ServerConf
	lastInvoke time.Time
	idleTime   time.Time
	isClosed   bool
	numInvoke  int32
}

//NewServer new Server and init with conf.
func NewServer(svr Protocol, conf *ServerConf) *Server {
	ts := &Server{svr: svr, conf: conf}
	ts.isClosed = false
	ts.lastInvoke = time.Now()
	return ts
}

func (ts *Server) getHandler() (sh ServerHandler) {
	if ts.conf.Proto == "tcp" {
		sh = &tcpHandler{conf: ts.conf, ts: ts}
	} else if ts.conf.Proto == "udp" {
		sh = &udpHandler{conf: ts.conf, ts: ts}
	}else if ts.conf.Proto == "ws" {
		sh = &websocketHandler{conf: ts.conf, ts: ts}
	} else {
		panic("unsupport protocol: " + ts.conf.Proto)
	}
	return
}

//Serve listen and handle
func (ts *Server) Serve() error {
	h := ts.getHandler()
	if err := h.Listen(); err != nil {
		return err
	}
	return h.Handle()
}

//Shutdown shutdown the server.
func (ts *Server) Shutdown() {
	ts.isClosed = true
}

//GetConfig gets the tars server config.
func (ts *Server) GetConfig() *ServerConf {
	return ts.conf
}

//IsZombie show whether the server is hanged by the request.
func (ts *Server) IsZombie(timeout time.Duration) bool {
	conf := ts.GetConfig()
	return conf.MaxInvoke != 0 && ts.numInvoke == conf.MaxInvoke && ts.lastInvoke.Add(timeout).Before(time.Now())
}

func (ts *Server) invoke(ctx context.Context, pkg []byte) {
	atomic.AddInt32(&ts.numInvoke, 1)
	ts.svr.Invoke(ctx, pkg)
	atomic.AddInt32(&ts.numInvoke, -1)
}
