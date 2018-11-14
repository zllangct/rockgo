package fserver

import (
	"fmt"
	"github.com/zllangct/RockGO/Network"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/timer"
	"github.com/zllangct/RockGO/utils"
	"net"
	"os"
	"os/signal"
	"time"
	"syscall"
)

func init() {
	utils.GlobalObject.Protoc = Network.NewProtocol()
	// --------------------------------------------init log start
	utils.ReSettingLog()
	// --------------------------------------------init log end
}

type Server struct {
	Name          string
	IPVersion     string
	IP            string
	Port          int
	MaxConn       int
	SessionIDPool chan uint32
	connectionMgr RockInterface.ISessionMgr
	Protoc        RockInterface.IServerProtocol
}

func (this *Server)GenSessionID(){
	//初始值
	count:=uint32(0)
	//初始化ID ConnPool
	this.SessionIDPool=make(chan uint32,100)
	//创建协程
	go func() {
		for{
			count++
			this.SessionIDPool<-count
		}
	}()
}

func NewServer() RockInterface.Iserver {
	s := &Server{
		Name:          utils.GlobalObject.Name,
		IPVersion:     "tcp4",
		IP:            "0.0.0.0",
		Port:          utils.GlobalObject.TcpPort,
		MaxConn:       utils.GlobalObject.MaxConn,
		connectionMgr: Network.NewConnectionMgr(),
		Protoc:        utils.GlobalObject.Protoc,
	}
	utils.GlobalObject.TcpServer = s

	return s
}

func NewTcpServer(name string, version string, ip string, port int, maxConn int, protoc RockInterface.IServerProtocol) RockInterface.Iserver {
	s := &Server{
		Name:          name,
		IPVersion:     version,
		IP:            ip,
		Port:          port,
		MaxConn:       maxConn,
		connectionMgr: Network.NewConnectionMgr(),
		Protoc:        protoc,
	}
	utils.GlobalObject.TcpServer = s

	return s
}

func (this *Server) handleConnection(conn *net.TCPConn) {
	sessionID :=<-this.SessionIDPool
	err:= conn.SetNoDelay(true)
	if err != nil {
		logger.Error(err)
	}
	err=conn.SetKeepAlive(true)
	if err != nil {
		logger.Error(err)
	}
	// conn.SetDeadline(time.Now().Add(time.Minute * 2))
	var fconn *Network.Session
	if this.Protoc == nil{
		fconn = Network.NewConnection(conn, sessionID, utils.GlobalObject.Protoc)

	}else{
		fconn = Network.NewConnection(conn, sessionID, this.Protoc)
	}
	fconn.SetProperty(Network.XINGO_CONN_PROPERTY_NAME, this.Name)
	fconn.Start()
}

func (this *Server) Start() {
	utils.GlobalObject.TcpServers[this.Name] = this
	go func() {
		this.Protoc.InitWorker(utils.GlobalObject.PoolSize)
		tcpAddr, err := net.ResolveTCPAddr(this.IPVersion, fmt.Sprintf("%s:%d", this.IP, this.Port))
		if err != nil{
			logger.Fatal("ResolveTCPAddr err: ", err)
			return
		}
		ln, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			logger.Error(err)
		}
		logger.Info(fmt.Sprintf("start xingo server %s...", this.Name))
		//构建会话ID
		this.GenSessionID()
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				logger.Error(err)
			}
			//设置链接存活时间
			//conn.SetDeadline(time.Now().Add(time.Second * 15))
			//max client exceed
			if this.connectionMgr.Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
			} else {
				go this.handleConnection(conn)
			}
		}
	}()
}

func (this *Server) GetConnectionMgr() RockInterface.ISessionMgr {
	return this.connectionMgr
}

func (this *Server) GetConnectionQueue() chan interface{} {
	return nil
}

func (this *Server) Stop() {
	logger.Info("stop xingo server ", this.Name)
	if utils.GlobalObject.OnServerStop != nil {
		utils.GlobalObject.OnServerStop()
	}
}

func (this *Server) AddRouter(router interface{}) {
	logger.Info("AddRouter")
	utils.GlobalObject.Protoc.GetMsgHandle().AddRouter(router)
}

func (this *Server) CallLater(durations time.Duration, f func(v ...interface{}), args ...interface{}) {
	delayTask := timer.NewTimer(durations, f, args)
	delayTask.Run()
}

func (this *Server) CallWhen(ts string, f func(v ...interface{}), args ...interface{}) {
	loc, err_loc := time.LoadLocation("Local")
	if err_loc != nil {
		logger.Error(err_loc)
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, loc)
	now := time.Now()
	if err == nil {
		if now.Before(t) {
			this.CallLater(t.Sub(now), f, args...)
		} else {
			logger.Error("CallWhen time before now")
		}
	} else {
		logger.Error(err)
	}
}

func (this *Server) CallLoop(durations time.Duration, f func(v ...interface{}), args ...interface{}) {
	go func() {
		delayTask := timer.NewTimer(durations, f, args)
		for {
			time.Sleep(delayTask.GetDurations())
			delayTask.GetFunc().Call()
		}
	}()
}

func (this *Server) WaitSignal() {
	signal.Notify(utils.GlobalObject.ProcessSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	sig := <-utils.GlobalObject.ProcessSignalChan
	logger.Info(fmt.Sprintf("server exit. signal: [%s]", sig))
	this.Stop()
}

func (this *Server) Serve() {
	this.Start()
	this.WaitSignal()
}
