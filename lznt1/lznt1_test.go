package lznt1

import (
	"crypto/sha1"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	source, err := os.ReadFile("../testdata/test.exe")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Compress(source)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(result))

	h := sha1.New()
	h.Write(result)
	t.Logf("%x", h.Sum(nil))

	if err = os.WriteFile("../testdata/go_lznt1_compressed.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}

func TestDecompress(t *testing.T) {
	source, err := os.ReadFile("../testdata/go_lznt1_compressed.bin")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Decompress(source)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(result))

	h := sha1.New()
	h.Write(result)
	t.Logf("%x", h.Sum(nil))

	if err = os.WriteFile("../testdata/go_lznt1_decompressed.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}
