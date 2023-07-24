package lznt1

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Decompressor struct {
	__input       []byte
	__inputCursor int
}

func (d *Decompressor) DecompressChunk(chunkLength int) ([]byte, error) {
	chunk := bytes.NewReader(d.__input[d.__inputCursor : d.__inputCursor+chunkLength])
	decompressed := &bytes.Buffer{}

	for chunk.Len() > 0 {
		flags, err := chunk.ReadByte()
		if err != nil {
			return nil, err
		}

		for i := 0; i < 8; i++ {
			if ((flags >> i) & 1) == 0 {
				if b, err := chunk.ReadByte(); err == nil {
					decompressed.WriteByte(b)
				} else {
					return nil, err
				}
			} else {
				pos := decompressed.Len() - 1
				mask := uint16(0x0fff)
				shift := uint16(12)

				for pos >= 0x10 {
					mask >>= 1
					shift--
					pos >>= 1
				}

				bs := make([]byte, 2)
				if _, err = chunk.Read(bs); err != nil {
					return nil, err
				}

				sym := binary.LittleEndian.Uint16(bs)
				length := (sym & mask) + 3
				offset := (sym >> shift) + 1
				index := decompressed.Len() - int(offset)

				if length >= offset {
					decompressed.Write(bytes.Repeat(decompressed.Bytes()[index:], 0xFFF/len(decompressed.Bytes()[index:])+1)[:length])
				} else {
					decompressed.Write(decompressed.Bytes()[index : index+int(length)])
				}
			}
			if chunk.Len() == 0 {
				break
			}
		}
	}

	return decompressed.Bytes(), nil
}

func (d *Decompressor) Decompress() ([]byte, error) {
	result := &bytes.Buffer{}

	for d.__inputCursor < len(d.__input) {
		// Read chunk header
		header := binary.LittleEndian.Uint16(d.__input[d.__inputCursor : d.__inputCursor+2])
		if header == 0 {
			return nil, fmt.Errorf("the compressed data is wrong")
		}

		chunkLength := int((header & 0x0FFF) + 1)

		if chunkLength >= len(d.__input) {
			return nil, fmt.Errorf("the compressed data is wrong")
		}

		d.__inputCursor += 2

		// Flags:
		//   Highest bit (0x8) means compressed
		// The other bits are always 011 (0x3) and have unknown meaning:
		//   The last two bits are possibly uncompressed chunk size (512, 1024, 2048, or 4096)
		//   However in NT 3.51, NT 4 SP1, XP SP2, Win 7 SP1 the actual chunk size is always 4096
		//   and the unknown flags are always 011 (0x3)

		if (header & 0x8000) != 0 {
			if decompressed, err := d.DecompressChunk(chunkLength); err == nil {
				result.Write(decompressed)
			} else {
				return nil, err
			}
		} else {
			result.Write(d.__input[d.__inputCursor : d.__inputCursor+chunkLength])
		}

		d.__inputCursor += chunkLength
	}

	return result.Bytes(), nil
}

func NewDecompressor(input []byte) *Decompressor {
	return &Decompressor{__input: input}
}
