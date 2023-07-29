package tcp

import (
	"bytes"
	"encoding/gob"
)

// gob协议
type GobSerializer struct {
}

func (g GobSerializer) Serialize(data any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (g GobSerializer) Deserialize(data []byte, target any) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(target)
}
