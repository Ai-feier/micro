package serialize

// Serializer rpc 序列化接口, 借鉴 go json 接口
type Serializer interface {
	Code() uint8
	Encode(val any) ([]byte, error)
	Decode(data []byte, val any) error
}