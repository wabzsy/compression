package main

import (
	"crypto/sha1"
	"flag"
	"github.com/wabzsy/compression"
	"log"
	"os"
	"time"
)

func main() {
	var input, output string
	var mode int
	flag.StringVar(&input, "i", "", "input file")
	flag.StringVar(&output, "o", "", "output file")
	flag.IntVar(&mode, "m", 0, `mode:
  1: aPLib Compress without header (golang)
  2: aPLib Compress with header (golang)
  3: aPLib Decompress with strict mode (golang)
  4: LZNT1 Compress (golang)
  5: LZNT1 Decompress (golang)
  6: RtlCompressBuffer (COMPRESSION_FORMAT_LZNT1 | COMPRESSION_ENGINE_MAXIMUM) -- Windows only
  7: RtlDecompressBuffer (COMPRESSION_FORMAT_LZNT1) -- Windows only
  8: Xpress Compress (golang)
  9: Xpress Decompress (golang)
  10: RtlCompressBuffer (COMPRESSION_FORMAT_XPRESS | COMPRESSION_ENGINE_MAXIMUM) -- Windows only
  11: RtlDecompressBuffer (COMPRESSION_FORMAT_XPRESS) -- Windows only
`)
	flag.Parse()

	if mode == 0 || input == "" || output == "" {
		flag.Usage()
		return
	}

	source, err := os.ReadFile(input)
	if err != nil {
		log.Fatalln(err)
	}

	var result []byte

	start := time.Now()
	defer func() {
		log.Println("elapsed time:", time.Since(start))
		log.Println("input length:", len(source))
		log.Printf("input sha1: %x\n", sha1Sum(source))
		log.Println("output length:", len(result))
		log.Printf("output sha1: %x\n", sha1Sum(result))
	}()

	switch mode {
	case 1:
		// aPLib Compress without header (golang)
		result, err = compression.APLibCompress(source)
	case 2:
		// aPLib Compress with header (golang)
		result, err = compression.APLibSafeCompress(source)
	case 3:
		// aPLib Decompress with strict mode (golang)
		result, err = compression.APLibStrictDecompress(source)
	case 4:
		// LZNT1 Compress (golang)
		result, err = compression.LZNT1Compress(source)
	case 5:
		// LZNT1 Decompress (golang)
		result, err = compression.LZNT1Decompress(source)
	case 6:
		// RtlCompressBuffer (COMPRESSION_FORMAT_LZNT1 | COMPRESSION_ENGINE_MAXIMUM)
		result, err = compression.RtlLZNT1Compress(source)
	case 7:
		// RtlDecompressBuffer (COMPRESSION_FORMAT_LZNT1)
		result, err = compression.RtlLZNT1Decompress(source)
	case 8:
		// Xpress Compress (golang)
		result, err = compression.XPressCompress(source)
	case 9:
		// Xpress Decompress (golang)
		result, err = compression.XPressDecompress(source)
	case 10:
		// RtlCompressBuffer (COMPRESSION_FORMAT_XPRESS | COMPRESSION_ENGINE_MAXIMUM)
		result, err = compression.RtlXPressCompress(source)
	case 11:
		// RtlDecompressBuffer (COMPRESSION_FORMAT_XPRESS)
		result, err = compression.RtlXPressDecompress(source)
	default:
		log.Fatalln("unknown mode")
	}

	if err != nil {
		log.Fatalln(err)
	}

	if err = os.WriteFile(output, result, 0666); err != nil {
		log.Fatalln(err)
	}
}

func sha1Sum(bs []byte) []byte {
	s := sha1.New()
	s.Write(bs)
	return s.Sum(nil)
}
