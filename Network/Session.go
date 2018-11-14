package Network

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"net"
	"sync"
	"time"
)

const (
	XINGO_CONN_PROPERTY_CTIME = "xingo_ctime"
	XINGO_CONN_PROPERTY_NAME = "xingo_tcpserver_name"
)

type Session struct {
	Conn         net.Conn
	isClosed     bool
	SessionId    uint32
	Protoc       RockInterface.IServerProtocol
	PropertyBag  map[string]interface{}
	sendtagGuard sync.RWMutex
	propertyLock sync.RWMutex

	SendBuffChan chan []byte
	ExtSendChan  chan bool
}

func NewConnection(conn net.Conn, sessionId uint32, protoc RockInterface.IServerProtocol) *Session {
	fconn := &Session{
		Conn:         conn,
		isClosed:     false,
		SessionId:    sessionId,
		Protoc:       protoc,
		PropertyBag:  make(map[string]interface{}),
		SendBuffChan: make(chan []byte, utils.GlobalObject.MaxSendChanLen),
		ExtSendChan:  make(chan bool, 1),
	}
	//set  connection time
	fconn.SetProperty(XINGO_CONN_PROPERTY_CTIME, time.Since(time.Now()))
	return fconn
}

func (this *Session) Start() {
	//add to connectionmsg
	serverName, err := this.GetProperty(XINGO_CONN_PROPERTY_NAME)
	if err != nil{
		logger.Error("not find server name in GlobalObject.")
		return
	}else{
		serverNameStr := serverName.(string)
		utils.GlobalObject.TcpServers[serverNameStr].GetConnectionMgr().Add(this)
	}

	this.Protoc.OnConnectionMade(this)
	this.startWriteThread()
	this.Protoc.StartReadThread(this)
}

func (this *Session) Stop() {
	// 防止将Send放在go内造成的多线程冲突问题
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if this.isClosed{
		return
	}
	 
	this.Conn.Close()
	this.ExtSendChan <- true
	this.isClosed = true
	//掉线回调放到go内防止，掉线回调处理出线死锁
	go this.Protoc.OnConnectionLost(this)
	//remove to connectionmsg
	serverName, err := this.GetProperty(XINGO_CONN_PROPERTY_NAME)
	if err != nil{
		logger.Error("not find server name in GlobalObject.")
		return
	}else{
		serverNameStr := serverName.(string)
		utils.GlobalObject.TcpServers[serverNameStr].GetConnectionMgr().Remove(this)
	}
	close(this.ExtSendChan)
	close(this.SendBuffChan)
}

func (this *Session) GetConnection() net.Conn {
	return this.Conn
}

func (this *Session) GetSessionId() uint32 {
	return this.SessionId
}

func (this *Session) GetProtoc() RockInterface.IServerProtocol {
	return this.Protoc
}

func (this *Session) GetProperty(key string) (interface{}, error) {
	this.propertyLock.RLock()
	defer this.propertyLock.RUnlock()

	value, ok := this.PropertyBag[key]
	if ok {
		return value, nil
	} else {
		return nil, errors.New("no property in connection")
	}
}

func (this *Session) SetProperty(key string, value interface{}) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	this.PropertyBag[key] = value
}

func (this *Session) RemoveProperty(key string) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	delete(this.PropertyBag, key)
}

func (this *Session) Send(data []byte) error {
	// 防止将Send放在go内造成的多线程冲突问题
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if !this.isClosed {
		if _, err := this.Conn.Write(data); err != nil {
			logger.Error(fmt.Sprintf("send data error.reason: %s", err))
			return err
		}
		return nil
	} else {
		return errors.New("connection closed")
	}
}

func (this *Session) SendBuff(data []byte) error {
	// 防止将Send放在go内造成的多线程冲突问题
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if !this.isClosed {

		// 发送超时
		select {
		case <-time.After(time.Second * 2):
			logger.Error("send error: timeout.")
			return errors.New("send error: timeout.")
		case this.SendBuffChan <- data:
			return nil
		}
	} else {
		return errors.New("connection closed")
	}

}

func (this *Session) RemoteAddr() net.Addr {
	return this.Conn.RemoteAddr()
}

func (this *Session) LostConnection() {
	this.Conn.Close()
	logger.Info("LostConnection session: ", this.SessionId)
}

func (this *Session) startWriteThread() {
	go func() {
		logger.Debug("start send data from channel...")
		for {
			select {
			case <-this.ExtSendChan:
				logger.Info("send thread exit successful!!!!")
				return
			case data := <-this.SendBuffChan:
				//send
				if _, err := this.Conn.Write(data); err != nil {
					logger.Info("send data error exit...")
					return
				}
			}
		}
	}()
}
