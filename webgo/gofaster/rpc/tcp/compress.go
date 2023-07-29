package tcp

type CompressInterface interface {
	Compress([]byte) ([]byte, error)
	UnCompress([]byte) ([]byte, error)
}

type CompressType byte

const (
	Gzip CompressType = iota
)
