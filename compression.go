package compression

import (
	"fmt"
	"github.com/wabzsy/compression/aplib"
	"github.com/wabzsy/compression/lznt1"
	"github.com/wabzsy/compression/rtl"
)

func APLibCompress(source []byte) ([]byte, error) {
	return aplib.Compress(source, false)
}

func APLibSafeCompress(source []byte) ([]byte, error) {
	return aplib.Compress(source, true)
}

func APLibDecompress(source []byte) ([]byte, error) {
	return aplib.Decompress(source, false)
}

func APLibStrictDecompress(source []byte) ([]byte, error) {
	return aplib.Decompress(source, true)
}

func LZNT1Compress(source []byte) ([]byte, error) {
	return lznt1.Compress(source)
}

func LZNT1Decompress(source []byte) ([]byte, error) {
	return lznt1.Decompress(source)
}

func XPressDecompress(source []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func XPressCompress(source []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func RtlLZNT1Compress(source []byte) ([]byte, error) {
	return rtl.LZNT1Compress(source)
}

func RtlLZNT1Decompress(source []byte) ([]byte, error) {
	return rtl.LZNT1Decompress(source)
}

func RtlXPressCompress(source []byte) ([]byte, error) {
	return rtl.XPressCompress(source)
}

func RtlXPressDecompress(source []byte) ([]byte, error) {
	return rtl.XPressDecompress(source)
}
