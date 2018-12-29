package MessageProtocol

import (
	"encoding/json"
)

type JsonProtocol struct {}

func NewJsonProtocol() *JsonProtocol {
	return &JsonProtocol{}
}

func (this *JsonProtocol)Marshal(message interface{})([]byte,error)  {
	return json.Marshal(message)
}

func (this *JsonProtocol)Unmarshal(data []byte, messageType interface{}) error {
	return json.Unmarshal(data,&messageType)
}