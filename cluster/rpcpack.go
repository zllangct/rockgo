package cluster

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/Network"
	"github.com/zllangct/RockGO/RockInterface"
	"encoding/gob"
	"github.com/zllangct/RockGO/logger"
)

type RpcData struct {
	MsgType RpcSignal              `json:"msgtype"`
	Key     string                 `json:"key,omitempty"`
	Target  string                 `json:"target,omitempty"`
	Args    []interface{}          `json:"args,omitempty"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

type RpcPackege struct {
	Len  int32
	Data []byte
}

type RpcRequest struct {
	Fconn   RockInterface.IWriter
	Rpcdata *RpcData
}

type RpcDataPack struct{}

func NewRpcDataPack() *RpcDataPack {
	return &RpcDataPack{}
}

func (this *RpcDataPack) GetHeadLen() int32 {
	return 4
}

func (this *RpcDataPack) Unpack(headdata []byte) (interface{}, error) {
	headbuf := bytes.NewReader(headdata)

	rp := &RpcPackege{}

	// 读取Len
	if err := binary.Read(headbuf, binary.LittleEndian, &rp.Len); err != nil {
		return nil, err
	}

	// 封包太大
	if rp.Len > Network.MaxPacketSize {
		return nil, errors.New("rpc packege too big!!!")
	}

	return rp, nil
}


func (this *RpcDataPack) Pack(msgId uint32, pkg interface{}) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 进行编码
	databuff := bytes.NewBuffer([]byte{})
	data := pkg.(*RpcData)
	if data != nil {
		enc := gob.NewEncoder(databuff)
		err = enc.Encode(data)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("rpcpack gob marshaling error:  %s", err))
		return
	}
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(databuff.Len())); err != nil {
		return
	}

	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, databuff.Bytes()); err != nil {
		return
	}

	out = outbuff.Bytes()
	return

}

func (this *RpcDataPack) Pack_byte(msgId uint32, data []byte) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(len(data))); err != nil {
		return
	}
	// 写MsgId
	if err = binary.Write(outbuff, binary.LittleEndian, msgId); err != nil {
		return
	}
	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, data); err != nil {
		return
	}
	out = outbuff.Bytes()
	return
}