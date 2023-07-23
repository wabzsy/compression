package rtl

import (
	"os"
	"testing"
)

func TestLZNT1Compress(t *testing.T) {
	source, err := os.ReadFile("big.exe")
	if err != nil {
		t.Fatal(err)
	}

	t.Skip()
	result, err := LZNT1Compress(source)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("compressed_lznt1.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}

func TestXPressCompress(t *testing.T) {
	source, err := os.ReadFile("big.exe")
	if err != nil {
		t.Fatal(err)
	}

	result, err := XPressCompress(source)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("compressed_xpress.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}

func TestLZNT1Decompress(t *testing.T) {
	source, err := os.ReadFile("./lznt1/go_compressed_lznt1.bin")
	if err != nil {
		t.Fatal(err)
	}

	result, err := LZNT1Decompress(source)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("./lznt1/go_decompressed_lznt1.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}

func TestXPressDecompress(t *testing.T) {
	source, err := os.ReadFile("compressed_xpress.bin")
	if err != nil {
		t.Fatal(err)
	}

	result, err := XPressDecompress(source)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("decompressed_xpress.bin", result, 0666); err != nil {
		t.Fatal(err)
	}
}
