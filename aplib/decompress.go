package aplib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type Decompressor struct {
	header      AP32Header
	reader      *bytes.Reader
	source      []byte
	destination *bytes.Buffer
	tag         uint8
	bitCount    int8
}

func (d *Decompressor) GetBit() uint8 {
	// check if tag is empty
	d.bitCount -= 1
	if d.bitCount < 0 {
		// load next tag
		d.tag = d.mustReadByte()
		d.bitCount = 7
	}

	// shift a bit out of tag
	bit := d.tag >> 7 & 1
	d.tag <<= 1

	return bit
}

func (d *Decompressor) GetGamma() int {
	result := 1
	// input gamma2-encoded bits
	for {
		result = (result << 1) + int(d.GetBit())
		if d.GetBit() == 0 {
			break
		}
	}

	return result
}

func (d *Decompressor) mustReadByte() uint8 {
	if b, err := d.reader.ReadByte(); err == nil {
		return b
	} else {
		panic(err)
	}
}

func (d *Decompressor) dePack() []byte {
	d.reader = bytes.NewReader(d.source)
	r0 := -1
	lwm := 0
	done := false

	// first byte verbatim
	d.destination.WriteByte(d.mustReadByte())

	// main decompression loop
	for !done {
		if d.GetBit() == 1 {
			if d.GetBit() == 1 {
				if d.GetBit() == 1 { // 1 1 1 singleByte
					offset := 0
					for i := 0; i < 4; i++ {
						offset = (offset << 1) + int(d.GetBit())
					}

					if offset != 0 {
						d.destination.WriteByte(d.destination.Bytes()[d.destination.Len()-offset])
					} else {
						d.destination.WriteByte(0)
					}

					lwm = 0
				} else { // short block 110
					offset := int(d.mustReadByte())
					length := 2 + (offset & 1)
					offset >>= 1

					if offset != 0 {
						for i := 0; i < length; i++ {
							d.destination.WriteByte(d.destination.Bytes()[d.destination.Len()-offset])
						}
					} else {
						done = true
					}

					r0 = offset
					lwm = 1
				}
			} else { // 1 0 block
				offset := d.GetGamma()

				if lwm == 0 && offset == 2 {
					offset = r0
					length := d.GetGamma()

					for i := 0; i < length; i++ {
						d.destination.WriteByte(d.destination.Bytes()[d.destination.Len()-offset])
					}
				} else {
					if lwm == 0 {
						offset -= 3
					} else {
						offset -= 2
					}

					offset <<= 8
					offset += int(d.mustReadByte())
					length := d.GetGamma()

					if offset >= 32000 {
						length++
					}

					if offset >= 1280 {
						length++
					}

					if offset < 128 {
						length += 2
					}

					for i := 0; i < length; i++ {
						d.destination.WriteByte(d.destination.Bytes()[d.destination.Len()-offset])
					}

					r0 = offset
				}

				lwm = 1
			}
		} else { // 0 literal
			d.destination.WriteByte(d.mustReadByte())
			lwm = 0
		}
	}

	return d.destination.Bytes()
}

func (d *Decompressor) Decompress(strict bool) ([]byte, error) {

	if bytes.HasPrefix(d.source, []byte("AP32")) && len(d.source) >= 24 {
		// data has an aPLib header
		if err := binary.Read(bytes.NewReader(d.source), binary.LittleEndian, &d.header); err != nil {
			return nil, err
		}
		d.source = d.source[d.header.HeaderSize : d.header.HeaderSize+d.header.PackedSize]
	}

	if strict {
		if d.header.PackedSize != 0 && int(d.header.PackedSize) != len(d.source) {
			return nil, fmt.Errorf("packed data size is incorrect")
		}
		if d.header.PackedCrc != 0 && d.header.PackedCrc != crc32.ChecksumIEEE(d.source) {
			return nil, fmt.Errorf("packed data checksum is incorrect")
		}
	}

	result := d.dePack()

	if strict {
		if d.header.OrigSize != 0 && int(d.header.OrigSize) != len(result) {
			return nil, fmt.Errorf("unpacked data size is incorrect")
		}
		if d.header.OrigCrc != 0 && d.header.OrigCrc != crc32.ChecksumIEEE(result) {
			return nil, fmt.Errorf("unpacked data checksum is incorrect")
		}
	}

	return result, nil
}

func NewDecompressor(source []byte) *Decompressor {
	return &Decompressor{
		source:      source,
		destination: &bytes.Buffer{},
		tag:         0,
		bitCount:    0,
	}
}
