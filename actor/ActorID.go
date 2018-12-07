package Actor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type ActorAddressType = int

var(
	//actor地址格式错误
    ErrActorWrongFormat = errors.New("this format is wrong,should be : IP:Port:LocalActorID")
)
/*
	ActorID such as "127.0.0.1:8888:0001",
	means "IP:PORT:LOCALACTORID"
*/
type ActorID  []string

func (this ActorID)String() string {
	buf:=bytes.Buffer{}
	for index, value := range this {
		buf.WriteString(value)
		if index !=len(this)-1{
			buf.WriteString(":")
		}
	}
	return buf.String()
}

//complate actor location id
func (this *ActorID) Parse(address string) (error){
	arr:= strings.Split(address,":")
	if len(arr)!=3{
		return ErrActorWrongFormat
	}
	*this = arr
	return nil
}

//get node address
func (this ActorID)GetNodeName() string  {
	return fmt.Sprintf("%s:%s",this[0],this[1])
}

//get node address
func (this ActorID)GetLocalActorID() string  {
	return this[2]
}