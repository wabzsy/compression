package lznt1

func Compress(input []byte) ([]byte, error) {
	return NewCompressor(input).Compress()
}

func Decompress(input []byte) ([]byte, error) {
	return NewDecompressor(input).Decompress()
}
