# Compression

[English](README.md) | [中文](README_zh.md)

`Compression` is a compression (decompression) library implemented in pure Go. It is mainly used to support some non-mainstream compression formats (such as: aPLib, LZNT1, XPRESS). It can be used as a module or as a stand-alone CLI tool.

## Directory Structure

| Directory | Description                                                  |
| --------- | ------------------------------------------------------------ |
| aplib     | Process data in aPLib format, support aPLib header           |
| lznt1     | Process data in COMPRESSION_FORMAT_LZNT1 format of RtlCompressBuffer |
| xpress    | Process data in COMPRESSION_FORMAT_XPRESS format of RtlCompressBuffer |
| rtl       | Use syscall to call the compression (decompression) function in ntdll.dll, **only supported on Windows platform** |
| example   | A simple CLI tool, see below for usage                       |
| testdata  | Empty                                                        |

## Usage

### Use as module

```bash
go get -v -u github.com/wabzsy/compression
```

```go
package example

import "github.com/wabzsy/compression"

func example() {
	input := []byte("abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc")

	// aPLib Compress without header (golang)
	result, err := compression.APLibCompress(input)
	if err != nil {
		panic(err)
	}

	// aPLib Compress with header (golang)
	result, err = compression.APLibSafeCompress(input)
	if err != nil {
		panic(err)
	}

	// aPLib Decompress with strict mode (golang)
	result, err = compression.APLibStrictDecompress(input)
	if err != nil {
		panic(err)
	}

	// LZNT1 Compress (golang)
	result, err = compression.LZNT1Compress(input)
	if err != nil {
		panic(err)
	}

	// LZNT1 Decompress (golang)
	result, err = compression.LZNT1Decompress(input)
	if err != nil {
		panic(err)
	}

	// XPRESS Compress (golang)
	result, err = compression.XPressCompress(input)
	if err != nil {
		panic(err)
	}

	// XPRESS Decompress (golang)
	result, err = compression.XPressDecompress(input)
	if err != nil {
		panic(err)
	}

	// RtlCompressBuffer (COMPRESSION_FORMAT_LZNT1 | COMPRESSION_ENGINE_MAXIMUM) -- Windows only
	result, err = compression.RtlLZNT1Compress(input)
	if err != nil {
		panic(err)
	}

	// RtlDecompressBuffer (COMPRESSION_FORMAT_LZNT1) -- Windows only
	result, err = compression.RtlLZNT1Decompress(input)
	if err != nil {
		panic(err)
	}

	// RtlCompressBuffer (COMPRESSION_FORMAT_XPRESS | COMPRESSION_ENGINE_MAXIMUM) -- Windows only
	result, err = compression.RtlXPressCompress(input)
	if err != nil {
		panic(err)
	}

	// RtlDecompressBuffer (COMPRESSION_FORMAT_XPRESS)  -- Windows only
	result, err = compression.RtlXPressDecompress(input)
	if err != nil {
		panic(err)
	}
}

```

### Use as a CLI tool

- clone the source code and compile:

```bash
git clone https://github.com/wabzsy/compression

cd compression/example

go build -v -o cli
```

- show `cli` usage:

```bash
./cli -h
```

```
Usage of ./cli:
  -i string
        input file
  -m int
        mode:
          1: aPLib Compress without header (golang)
          2: aPLib Compress with header (golang)
          3: aPLib Decompress with strict mode (golang)
          4: LZNT1 Compress (golang)
          5: LZNT1 Decompress (golang)
          6: RtlCompressBuffer (COMPRESSION_FORMAT_LZNT1 | COMPRESSION_ENGINE_MAXIMUM)
          7: RtlDecompressBuffer (COMPRESSION_FORMAT_LZNT1)
          8: XPRESS Compress (golang)
          9: XPRESS Decompress (golang)
          10: RtlCompressBuffer (COMPRESSION_FORMAT_XPRESS | COMPRESSION_ENGINE_MAXIMUM)
          11: RtlDecompressBuffer (COMPRESSION_FORMAT_XPRESS)
        
  -o string
        output file
```

- example of input/output:

```bash
# XPRESS Compress 
./cli -i ../testdata/test.exe -o ../testdata/test.bin -m 8
```

```
2023/07/26 06:59:19 elapsed time: 1.217532875s
2023/07/26 06:59:19 input length: 7905792
2023/07/26 06:59:19 input sha1: 02584ea42efe09e83e9093e1e76ec319930a55c3
2023/07/26 06:59:19 output length: 3833366
2023/07/26 06:59:19 output sha1: 0a07eb4e0ddac9864125f319eaac43488bef90e4
```

```bash
# XPRESS Decompress 
./cli -i ../testdata/test.bin -o ../testdata/test.dec -m 9
```

```
2023/07/26 07:00:33 elapsed time: 56.612625ms
2023/07/26 07:00:33 input length: 3833366
2023/07/26 07:00:33 input sha1: 0a07eb4e0ddac9864125f319eaac43488bef90e4
2023/07/26 07:00:33 output length: 7905792
2023/07/26 07:00:33 output sha1: 02584ea42efe09e83e9093e1e76ec319930a55c3
```

## References & Links

https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-xca/a8b7cb0a-92a6-4187-a23b-5e14273b96f8

https://ibsensoftware.com/products_aPLib.html

https://github.com/coderforlife/ms-compress C++

https://github.com/emmanuel-marty/apultra - C

https://github.com/li-xilin/lznt1 - C

https://github.com/svendahl/cap - C#

https://github.com/you0708/lznt1 - Python

https://github.com/nma-io/refinery - Python

https://github.com/snemes/kabopan - Python

https://github.com/snemes/aplib - Python

https://github.com/herrcore/aplib-ripper - Python

https://github.com/CERT-Polska/malduck - Python

https://github.com/killeven/lznt1 - Golang

https://github.com/hatching/aplib - Golang

https://github.com/julyanserra/Basic-LZ77-in-Python

https://github.com/SirusDoma/klz77

https://github.com/fbonhomm/LZ77

