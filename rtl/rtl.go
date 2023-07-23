//go:build !windows

package rtl

import "fmt"

var (
	ERR_BAD_OS = fmt.Errorf("the 'compression/rtl' package is not supported on this operating system")
)

func LZNT1Compress(source []byte) ([]byte, error) {
	return nil, ERR_BAD_OS
}

func XPressCompress(source []byte) ([]byte, error) {
	return nil, ERR_BAD_OS
}

func LZNT1Decompress(source []byte) ([]byte, error) {
	return nil, ERR_BAD_OS
}

func XPressDecompress(source []byte) ([]byte, error) {
	return nil, ERR_BAD_OS
}
