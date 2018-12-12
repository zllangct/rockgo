package Actor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var(
	//actor地址格式错误
    ErrActorWrongFormat = errors.New("this format is wrong,should be : IP:Port:LocalActorID")
)

type ActorAddressType = int

/*
	Target such as "127.0.0.1:8888:XXXXXXX",
	means "IP:PORT:LOCALATORID"
*/
type ActorID  []string



func EmptyActorID() ActorID {
	return make([]string,3)
}

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

func (this *ActorID) SetNodeID(address string) (ActorID,error) {
	arr:= strings.Split(address,":")
	if len(arr)!=2{
		return nil,ErrActorWrongFormat
	}
	(*this)[0]=arr[0]
	(*this)[1]=arr[1]
	return *this,nil
}

func (this *ActorID) SetLocalActorID(id string) *ActorID{
	(*this)[2]=id
	return this
}

func (this *ActorID) Parse(address string) (error){
	arr:= strings.Split(address,":")
	if len(arr)!=3{
		return ErrActorWrongFormat
	}
	*this = arr
	return nil
}

//get node address
func (this ActorID) GetNodeID() string  {
	return fmt.Sprintf("%s:%s",this[0],this[1])
}

//get node address
func (this ActorID)GetLocalActorID() string  {
	return this[2]
}