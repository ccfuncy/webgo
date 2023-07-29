package tcp

type Serializer interface {
	Serialize(data any) ([]byte, error)
	Deserialize(data []byte, target any) error
}

type SerializeType byte

const (
	Gob SerializeType = iota
	ProtoBuff
)
