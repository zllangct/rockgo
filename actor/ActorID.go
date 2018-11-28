package Actor

import (
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
	//actor地址不完整
	ERR_ID_NOT_COMPLETE = errors.New("actor id is not complete")
)
/*
	ActorID such as "0001:0025:2214:0001:0001",
	means "APPID:RULE:NODE:GROUP:LOCALACTORID"a
*/
type ActorID  string

func (this ActorID)ToString() string {
	return string(this)
}
//local actor id
func (this ActorID)LocalID() ([]string,error) {
	s:= strings.Split(string(this),":")
	if len(s)<2 {
		return s, ERR_ID_NOT_COMPLETE
	}
	return s[len(s)-2:],nil
}
//complate actor location id
func (this ActorID)ToArray() []string {
	return strings.Split(string(this),":")
}
//returns address of actor
func (this ActorID)GetAddress(addressType ActorAddressType) string  {
	return  strings.Split(string(this),":")[addressType]
}