package network

import (
	"context"
	"encoding/binary"
	"fmt"
)
/*  LTD protocol
	Length—（Type—Data） ，数据长度—（消息类型—消息体） 大小：  4 — （4 — n）
*/
type LtdProtocol struct{}

//Invoke recv request and make response.
func (s *LtdProtocol) Invoke(ctx context.Context,req []byte){
	fmt.Println("recv", string(req))

}

//ParsePackage parse package from buff,check if tars package finished.
func (s *LtdProtocol) ParsePackage(buff []byte) (pkgLen, status int) {
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