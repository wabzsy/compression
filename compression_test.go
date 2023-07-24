package compression

import (
	"crypto/sha1"
	"os"
	"path"
	"testing"
)

func TestAPLibCompress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "go_aplib_compressd"),
		APLibCompress,
	)
}

func TestAPLibDecompress(t *testing.T) {
	run(t,
		path.Join("testdata", "go_aplib_compressd"),
		path.Join("testdata", "go_aplib_decompressd"),
		APLibDecompress,
	)
}

func TestAPLibSafeCompress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "go_aplib_safe_compressd"),
		APLibCompress,
	)
}

func TestAPLibStrictDecompress(t *testing.T) {
	run(t,
		path.Join("testdata", "go_aplib_compressd"),
		path.Join("testdata", "go_aplib_decompressd"),
		APLibStrictDecompress,
	)
}

func TestLZNT1Compress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "go_lznt1_compressd"),
		LZNT1Compress,
	)
}

func TestLZNT1Decompress(t *testing.T) {
	run(t,
		path.Join("testdata", "go_lznt1_compressd"),
		path.Join("testdata", "go_lznt1_decompressd"),
		LZNT1Decompress,
	)
}

func TestRtlLZNT1Compress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "rtl_lznt1_compressd"),
		RtlLZNT1Compress,
	)
}

func TestRtlLZNT1Decompress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "rtl_lznt1_compressd"),
		RtlLZNT1Decompress,
	)
}

func TestRtlXPressCompress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "rtl_xpress_compressd"),
		RtlXPressCompress,
	)
}

func TestRtlXPressDecompress(t *testing.T) {
	run(t,
		path.Join("testdata", "rtl_xpress_compressd"),
		path.Join("testdata", "rtl_xpress_decompressd"),
		RtlXPressDecompress,
	)
}

func TestXPressCompress(t *testing.T) {
	run(t,
		path.Join("testdata", "test.exe"),
		path.Join("testdata", "go_xpress_compressd"),
		XPressCompress,
	)
}

func TestXPressDecompress(t *testing.T) {
	run(t,
		path.Join("testdata", "go_xpress_compressd"),
		path.Join("testdata", "go_xpress_decompressd"),
		XPressDecompress,
	)
}

func sha1Sum(bs []byte) []byte {
	s := sha1.New()
	s.Write(bs)
	return s.Sum(nil)
}

func run(t *testing.T, inputFile, outputFile string, fn func([]byte) ([]byte, error)) {
	source, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("input length:", len(source))
	t.Logf("input hash: %x\n", sha1Sum(source))

	result, err := fn(source)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("output length:", len(result))
	t.Logf("output hash: %x", sha1Sum(result))

	if err = os.WriteFile(outputFile, result, 0666); err != nil {
		t.Fatal(err)
	}
}
