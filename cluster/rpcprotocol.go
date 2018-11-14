package cluster

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"io"
)

type RpcServerProtocol struct {
	rpcMsgHandle *RpcMsgHandle
	rpcDatapack  *RpcDataPack
}

func NewRpcServerProtocol() *RpcServerProtocol {
	return &RpcServerProtocol{
		rpcMsgHandle: NewRpcMsgHandle(),
		rpcDatapack:  NewRpcDataPack(),
	}
}

func (this *RpcServerProtocol) GetMsgHandle() RockInterface.Imsghandle {
	return this.rpcMsgHandle
}

func (this *RpcServerProtocol) GetDataPack() RockInterface.Idatapack {
	return this.rpcDatapack
}

func (this *RpcServerProtocol) AddRpcRouter(router interface{}) {
	this.rpcMsgHandle.AddRouter(router)
}

func (this *RpcServerProtocol) InitWorker(poolsize int32) {
	this.rpcMsgHandle.StartWorkerLoop(int(poolsize))
}

func (this *RpcServerProtocol) OnConnectionMade(fconn RockInterface.ISession) {
	logger.Info(fmt.Sprintf("node ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterConnectioned(fconn)
}

func (this *RpcServerProtocol) OnConnectionLost(fconn RockInterface.ISession) {
	logger.Info(fmt.Sprintf("node ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterClosed(fconn)
}

func (this *RpcServerProtocol) StartReadThread(fconn RockInterface.ISession) {
	logger.Debug("start receive rpc data from socket...")
	var state int = 1   //读取状态,1:读取包头,2:读取包体
	var pkg *RpcPackege //包数据
	for {
		//读取包头
		if state == 1 {
			//read per head data
			headdata := make([]byte, this.rpcDatapack.GetHeadLen())
			if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
				if err == io.EOF || err ==io.ErrUnexpectedEOF{
					continue
				}else{
					logger.Error("断开链接:"+err.Error())
					fconn.Stop()
					return
				}
				//logger.Error("断开链接:"+err.Error())
				//fconn.Stop()
				//return
			}
			pkgOnlyHead, err := this.rpcDatapack.Unpack(headdata)
			if err != nil {
				logger.Error("断开链接:"+err.Error())
				fconn.Stop()
				return
			}
			pkg = pkgOnlyHead.(*RpcPackege)
			if !(pkg.Len > 0) {
				logger.Error("DD 1")
				continue
			}
			state = 2
		}
		//读取包体
		if state == 2 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				if err == io.EOF || err ==io.ErrUnexpectedEOF{
					continue
				}else{
					logger.Error("断开链接:"+err.Error())
					fconn.Stop()
					return
				}
				//logger.Error("断开链接:"+err.Error())
				//fconn.Stop()
				//return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}
				dec := gob.NewDecoder(bytes.NewReader(pkg.Data))
				err = dec.Decode(rpcRequest.Rpcdata)
				if err != nil {
					logger.Error(err)
					fconn.Stop()
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, int(rpcRequest.Rpcdata.MsgType)))
				if utils.GlobalObject.PoolSize > 0 && rpcRequest.Rpcdata.MsgType != RESPONSE {
					this.rpcMsgHandle.DeliverToMsgQueue(rpcRequest)
				} else {
					this.rpcMsgHandle.DoMsgFromGoRoutine(rpcRequest)
				}
			}
			state=1
		}
	}
}

type RpcClientProtocol struct {
	rpcMsgHandle *RpcMsgHandle
	rpcDatapack  *RpcDataPack
}

func NewRpcClientProtocol() *RpcClientProtocol {
	return &RpcClientProtocol{
		rpcMsgHandle: NewRpcMsgHandle(),
		rpcDatapack:  NewRpcDataPack(),
	}
}

func (this *RpcClientProtocol) GetMsgHandle() RockInterface.Imsghandle {
	return this.rpcMsgHandle
}

func (this *RpcClientProtocol) GetDataPack() RockInterface.Idatapack {
	return this.rpcDatapack
}
func (this *RpcClientProtocol) AddRpcRouter(router interface{}) {
	this.rpcMsgHandle.AddRouter(router)
}

func (this *RpcClientProtocol) InitWorker(poolsize int32) {
	this.rpcMsgHandle.StartWorkerLoop(int(poolsize))
}

func (this *RpcClientProtocol) OnConnectionMade(fconn RockInterface.Iclient) {
	utils.GlobalObject.OnClusterCConnectioned(fconn)
}

func (this *RpcClientProtocol) OnConnectionLost(fconn RockInterface.Iclient) {
	utils.GlobalObject.OnClusterCClosed(fconn)
}

func (this *RpcClientProtocol) StartReadThread(fconn RockInterface.Iclient) {
	logger.Debug("start receive rpc data from socket...")
	var state int = 1   //读取状态,1:读取包头,2:读取包体
	var pkg *RpcPackege =&RpcPackege{}
	for {
		//读取包头
		if state == 1 {
			//read per head data
			headdata := make([]byte, this.rpcDatapack.GetHeadLen())
			if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
				if err == io.EOF || err ==io.ErrUnexpectedEOF{
					continue
				}else{
					logger.Error("断开链接:"+err.Error())
					fconn.Stop(false)
					return
				}
				//logger.Error("断开链接:"+err.Error())
				//fconn.Stop(false)
				//return
			}
			pkgOnlyHead, err := this.rpcDatapack.Unpack(headdata)
			if err != nil {
				logger.Error("断开链接:"+err.Error())
				fconn.Stop(false)
				return
			}
			pkg = pkgOnlyHead.(*RpcPackege)
			if !(pkg.Len > 0) {
				logger.Error("DD 1")
				continue
			}
			state = 2
		}
		//读取包体
		if state == 2 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				if err == io.EOF || err ==io.ErrUnexpectedEOF{
					continue
				}else{
					logger.Error("断开链接:"+err.Error())
					fconn.Stop(false)
					return
				}
				//logger.Error("断开链接:"+err.Error())
				//fconn.Stop(false)
				//return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}
				dec := gob.NewDecoder(bytes.NewReader(pkg.Data))
				err = dec.Decode(rpcRequest.Rpcdata)
				if err != nil {
					logger.Error(err)
					fconn.Stop(false)
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, int(rpcRequest.Rpcdata.MsgType)))
				if utils.GlobalObject.PoolSize > 0 && rpcRequest.Rpcdata.MsgType != RESPONSE {
					this.rpcMsgHandle.DeliverToMsgQueue(rpcRequest)
				} else {
					this.rpcMsgHandle.DoMsgFromGoRoutine(rpcRequest)
				}
			}
			state=1
		}
	}
}