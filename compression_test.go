package compression

import (
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"
)

func TestAPLibCompress(t *testing.T) {
	run(t,
		"test.exe",
		"go_aplib_compressd",
		APLibCompress,
	)
}

func TestAPLibDecompress(t *testing.T) {
	run(t,
		"go_aplib_compressd",
		"go_aplib_decompressd",
		APLibDecompress,
	)
}

func TestAPLibSafeCompress(t *testing.T) {
	run(t,
		"test.exe",
		"go_aplib_safe_compressd",
		APLibCompress,
	)
}

func TestAPLibStrictDecompress(t *testing.T) {
	run(t,
		"go_aplib_compressd",
		"go_aplib_decompressd",
		APLibStrictDecompress,
	)
}

func TestLZNT1Compress(t *testing.T) {
	run(t,
		"test.exe",
		"go_lznt1_compressd",
		LZNT1Compress,
	)
}

func TestLZNT1Decompress(t *testing.T) {
	run(t,
		"go_lznt1_compressd",
		"go_lznt1_decompressd",
		LZNT1Decompress,
	)
}

func TestRtlLZNT1Compress(t *testing.T) {
	run(t,
		"test.exe",
		"rtl_lznt1_compressd",
		RtlLZNT1Compress,
	)
}

func TestRtlLZNT1Decompress(t *testing.T) {
	run(t,
		"test.exe",
		"rtl_lznt1_compressd",
		RtlLZNT1Decompress,
	)
}

func TestRtlXPressCompress(t *testing.T) {
	run(t,
		"test.exe",
		"rtl_xpress_compressd",
		RtlXPressCompress,
	)
}

func TestRtlXPressDecompress(t *testing.T) {
	run(t,
		"rtl_xpress_compressd",
		"rtl_xpress_decompressd",
		RtlXPressDecompress,
	)
}

func TestXPressCompress(t *testing.T) {
	run(t,
		"test.exe",
		"go_xpress_compressd",
		XPressCompress,
	)
}

func TestXPressDecompress(t *testing.T) {
	run(t,
		"go_xpress_compressd",
		"go_xpress_decompressd",
		XPressDecompress,
	)
}

func sha1Sum(bs []byte) []byte {
	s := sha1.New()
	s.Write(bs)
	return s.Sum(nil)
}

func run(t *testing.T, inputFile, outputFile string, fn func([]byte) ([]byte, error)) {
	source, err := os.ReadFile(filepath.Join("testdata", inputFile))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("input size:", len(source))
	t.Logf("input SHA1: %x\n", sha1Sum(source))

	result, err := fn(source)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("output size:", len(result))
	t.Logf("output SHA1: %x", sha1Sum(result))

	if err = os.WriteFile(filepath.Join("testdata", outputFile), result, 0666); err != nil {
		t.Fatal(err)
	}
}
