package xpress

func CompressWithLevel(input []byte, level Level) ([]byte, error) {
	return NewCompressor(input, level).Compress()
}

func Compress(input []byte) ([]byte, error) {
	// Level3的行为和RTL最相近(时间/压缩率), 但是当前使用场景对时间要求不高, 所以暂时默认用7
	// https://github.com/coderforlife/ms-compress/issues/3
	return CompressWithLevel(input, Level7)
}

func Decompress(source []byte) ([]byte, error) {
	return NewDecompressor(source).Decompress()
}
