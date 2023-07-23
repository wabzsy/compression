package aplib

type AP32Header struct {
	Magic      [4]byte
	HeaderSize uint32
	PackedSize uint32
	PackedCrc  uint32
	OrigSize   uint32
	OrigCrc    uint32
}

func Decompress(input []byte, strict bool) ([]byte, error) {
	return NewDecompressor(input).Decompress(strict)
}

func Compress(input []byte, safe bool) ([]byte, error) {
	return NewCompressor(input).Compress(safe)
}
