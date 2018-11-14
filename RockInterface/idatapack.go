package RockInterface

type Idatapack interface {
	GetHeadLen() int32
	Unpack([]byte) (interface{}, error)
	Pack(uint32, interface{}) ([]byte, error)
	Pack_byte(uint32, []byte) ([]byte, error)
}
