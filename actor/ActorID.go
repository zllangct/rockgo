package Actor

import (
	"bytes"
	"errors"
	"strings"
)

type ActorAddressType = int

const (
	ADRESS_APPID = iota
	ADRESS_RULE
	ADRESS_GROUP
	ADDRESS_NODE
	ADRESS_LOCALACTORID
)
var(
	//actor地址格式错误
    ErrActorWrongFormat = errors.New("this format is wrong,should be : xx:cc:vv:bb:nn")
)
/*
	ActorID such as "0001:0025:2214:0001:0001",
	means "APPID:RULE:NODE:GROUP:LOCALACTORID"
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
	if len(arr)!=5{
		return ErrActorWrongFormat
	}
	*this = arr
	return nil
}

//returns address of actor
func (this ActorID)GetSeparation(addressType ActorAddressType) string  {
	return  this[addressType]
}