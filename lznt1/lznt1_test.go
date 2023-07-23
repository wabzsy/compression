package lznt1

import (
	"crypto/sha1"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {

	source, err := os.ReadFile("../WindowsCodecsRaw.dll")
	if err != nil {
		t.Fatal(err)
	}

	result, err := NewCompressor(source).Compress()
	//result := Compress(source)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(result))

	h := sha1.New()
	h.Write(result)
	t.Logf("%x", h.Sum(nil))

	if err = os.WriteFile("go_compressed_lznt1.bin", result, 0666); err != nil {
		t.Fatal(err)
	}

}
