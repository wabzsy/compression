package xpress

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	ErrInvalidData = fmt.Errorf("the input data is invalid")
)

type Decompressor struct {
	__input       []byte
	__inputCursor int

	reader *bytes.Reader
	output *bytes.Buffer
}

func (d *Decompressor) ReadUint32() (uint32, error) {
	bs := make([]byte, 4)
	if _, err := d.reader.Read(bs); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bs), nil
}

func (d *Decompressor) ReadUint16() (uint16, error) {
	bs := make([]byte, 2)
	if _, err := d.reader.Read(bs); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(bs), nil
}

func (d *Decompressor) literal() error {
	if c, err := d.ReadByte(); err == nil {
		return d.output.WriteByte(c)
	} else {
		return err
	}
}

func (d *Decompressor) ReadByte() (uint8, error) {
	return d.reader.ReadByte()
}

// SetBitsAreHighest This function checks that a number is of the form 1..10..0 -
// basically all 1 bits are more significant than all 0 bits
// (also allowed are all 1s and all 0s)
func SetBitsAreHighest(x uint32) bool {
	x = ^x
	return ((x + 1) & x) == 0
}

func (d *Decompressor) Decompress() ([]byte, error) {
	d.reader = bytes.NewReader(d.__input)
	d.output = &bytes.Buffer{}

	var halfByte *uint8

	for d.reader.Len()-4 > 0 {
		flags, err := d.ReadUint32()
		if err != nil {
			return nil, err
		}
		flagged := flags & 0x80000000
		flags = (flags << 1) | 1
		for {
			// Either: offset/length symbol, end of flags, or end of stream (checked above)
			if flagged != 0 {
				// 以下
				symbol, err := d.ReadUint16()
				if err != nil {
					return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Unable to read 2 bytes for offset/length")
				}

				offset := (symbol >> 3) + 1
				length := uint32(symbol & 0x7)

				if length == 0x7 {
					if halfByte != nil {
						length = uint32(*halfByte >> 4)
						halfByte = nil
					} else {
						if n, err := d.ReadByte(); err != nil {
							return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Unable to read a half-byte for length")
						} else {
							halfByte = &n
							length = uint32(*halfByte & 0xF)
						}
					}

					if length == 0xF {
						if n, err := d.ReadByte(); err != nil {
							return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Unable to read a byte for length")
						} else {
							length = uint32(n)
						}

						if length == 0xFF {
							if n, err := d.ReadUint16(); err != nil {
								return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Unable to read two bytes for length")
							} else {
								length = uint32(n)
							}

							if length == 0 {
								if length, err = d.ReadUint32(); err != nil {
									return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Unable to read four bytes for length")
								}
							}

							if length < 0xF+0x7 {
								return nil, fmt.Errorf("XPRESS Decompression Error: Invalid data: Invalid length")
							}
							length -= 0xF + 0x7
						}
						length += 0xF
					}
					length += 0x7
				}
				length += 0x3

				// 以上
				if d.output.Len()-int(offset) < 0 {
					return nil, ErrInvalidData
				}

				if offset == 1 {
					// out 当前偏移的前一位*长度 TODO: 优化
					c := d.output.Bytes()[d.output.Len()-1]
					for i := uint32(0); i < length; i++ {
						d.output.WriteByte(c)
					}
				} else {
					// 从out指定偏移位置复制一段长度为length的片段 TODO: 优化
					for i := uint32(0); i < length; i++ {
						d.output.WriteByte(d.output.Bytes()[d.output.Len()-int(offset)])
					}
				}
			} else {
				if err = d.literal(); err != nil {
					return nil, err
				}
			}

			flagged = flags & 0x80000000
			flags <<= 1

			if d.reader.Len() == 0 {
				// 检查异常
				if flagged == 0 || !SetBitsAreHighest(flags) {
					return nil, ErrInvalidData
				}
				// 返回结果
				return d.output.Bytes(), nil
			}

			if flags == 0 {
				break
			}
		}
	}

	return nil, ErrInvalidData
}

func NewDecompressor(input []byte) *Decompressor {
	return &Decompressor{
		__input: input,
	}
}
