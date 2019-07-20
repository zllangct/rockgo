package network

import (
	"context"
	"encoding/binary"
)

/*  LTD protocol
Length—（Session-Type—Data） ，数据长度—（会话—消息类型—消息体） 大小：  4 — （4—4 —n）
*/
type LstdProtocol struct{}

//ParseMessage recv request and make response.
func (s *LstdProtocol) ParseMessage(ctx context.Context, data []byte) ([]uint32, []byte) {
	sess := binary.BigEndian.Uint32(data[:4])
	mid := binary.BigEndian.Uint32(data[4:8])
	return []uint32{sess, mid}, data[8:]
}

//ParsePackage parse package from buff,check if tars package finished.
func (s *LstdProtocol) ParsePackage(buff []byte) (pkgLen, status int) {
	if len(buff) < 4 {
		return 0, PACKAGE_LESS
	}
	length := binary.BigEndian.Uint32(buff[:4])

	if length > 1048576000 || len(buff) > 1048576000 { // 1000MB
		return 0, PACKAGE_ERROR
	}
	if len(buff) < int(length) {
		return 0, PACKAGE_LESS
	}
	return int(length), PACKAGE_FULL
}
