package MessageProtocol

import (
	"github.com/golang/protobuf/proto"
)

type ProtobufProtocol struct{}

func NewProtobufProtocol() *ProtobufProtocol {
	return &ProtobufProtocol{}
}

func (this *ProtobufProtocol) Marshal(message interface{}) ([]byte, error) {
	return proto.Marshal(message.(proto.Message))
}

func (this *ProtobufProtocol) Unmarshal(data []byte, messageType interface{}) error {
	return proto.Unmarshal(data, messageType.(proto.Message))
}
