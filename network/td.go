package network

import (
	"context"
	"encoding/binary"
)
/*  TD protocol
	Type—Data ，消息类型—消息体 大小：  4 — n
*/
type TdProtocol struct{}

func (s *TdProtocol) ParseMessage(ctx context.Context,data []byte)([]uint32,[]byte){
	mt := binary.BigEndian.Uint32(data[:4])
	return []uint32{mt}, data[4:]
}

//不会用到
func (s *TdProtocol) ParsePackage(buff []byte) (pkgLen, status int) {
	return 0,0
} 