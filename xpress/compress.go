package xpress

import (
	"bytes"
	"encoding/binary"
	"math"
)

const (
	MAX_OFFSET = 0x2000
	CHUNK_SIZE = 0x2000
	HASH_BITS  = 15
)

type Level struct {
	MaxChain   int
	NiceLength int
}

var (
	Level1 = Level{NiceLength: 16, MaxChain: 4}
	Level2 = Level{NiceLength: 32, MaxChain: 8}
	Level3 = Level{NiceLength: 48, MaxChain: 11}
	Level4 = Level{NiceLength: 64, MaxChain: 16}
	Level5 = Level{NiceLength: 128, MaxChain: 32}
	Level6 = Level{NiceLength: 256, MaxChain: 64}
	Level7 = Level{NiceLength: 512, MaxChain: 128}
	Level8 = Level{NiceLength: math.MaxInt, MaxChain: math.MaxInt}
)

type Dictionary struct {
	level Level

	// Window properties
	WindowSize uint32
	WindowMask uint32

	// The hashing function, which works progressively
	HashSize  uint32
	HashMask  uint32
	HashShift uint32

	table  map[uint16]int
	window map[uint32]int
}

func (d *Dictionary) WindowPos(x int) uint32 {
	return uint32((x) & int(d.WindowMask))
}

func (d *Dictionary) HashUpdate(h uint16, c uint8) uint16 {
	// h 必须是uint16, 不然无法在左移时抹掉高位的溢出, 会导致计算错误
	return ((h << uint16(d.HashShift)) ^ uint16(c)) & uint16(d.HashMask)
}

func (d *Dictionary) Fill(data []byte, cursor int) int {
	if cursor >= len(data)-2 {
		return len(data) - 2
	}

	pos := d.WindowPos(cursor)
	end := len(data) - 2
	if cursor+CHUNK_SIZE < len(data)-2 {
		end = cursor + CHUNK_SIZE
	}

	hash := d.HashUpdate(uint16(data[cursor]), data[cursor+1])

	for cursor < end {
		hash = d.HashUpdate(hash, data[cursor+2])
		// 因为此处存的是偏移量, 为了区分偏移量的0和不存在的0, 所以要判断一下是否存在再存
		// C++版这里存的是指针地址, 所以不存在这个问题
		// d.window[pos] = d.table[hash]
		if v, ok := d.table[hash]; ok {
			d.window[pos] = v
		}
		pos++
		d.table[hash] = cursor
		cursor++
	}

	return end
}

func (d *Dictionary) Find(input []byte, cursor int) (length, offset int) {
	//length = 2
	position, ok := d.window[d.WindowPos(cursor)]
	chainLength := d.level.MaxChain

	//prefix := binary.LittleEndian.Uint16(input[cursor : cursor+2])

	for chainLength != 0 && ok && position >= cursor-MAX_OFFSET {
		if bytes.Equal(input[position:position+2], input[cursor:cursor+2]) {
			//i := 2
			i := 3
			// at this point the at least 3 bytes are matched (due to the hashing function forcing byte 3 to the same)
			// TODO: 此处还可继续优化, 比如强转uint32同时比较两位,不符后再比较最后一次的位置
			// TODO: 此处-1还是不-1还需要考虑一下
			//for cursor+i < len(input) -1 && input[cursor+i] == input[position+i] {
			for cursor+i < len(input) && input[cursor+i] == input[position+i] {
				i++
			}

			if i > length {
				offset = cursor - position
				length = i
				if length >= d.level.NiceLength {
					break
				}
			}
		}
		position, ok = d.window[d.WindowPos(position)]
		chainLength--
	}

	return length, offset
}

func NewDictionary(level Level) *Dictionary {
	d := &Dictionary{
		level:      level,
		WindowSize: CHUNK_SIZE << 1,
		HashSize:   1 << HASH_BITS,
		HashShift:  (HASH_BITS + 2) / 3,
		window:     make(map[uint32]int), // d.WindowSize
		table:      make(map[uint16]int), // d.HashSize
	}
	d.WindowMask = d.WindowSize - 1
	d.HashMask = d.HashSize - 1
	return d
}

type Compressor struct {
	dict           *Dictionary
	__input        []byte
	__inputCursor  int
	__output       []byte
	__outputCursor int

	__flagCount  int
	__flagCursor int
}

func (c *Compressor) MaxCompressedSize() int {
	return len(c.__input) + 4 + 4*(len(c.__input)/32)
}

func (c *Compressor) __SetFlags(flags uint32) {
	c.__flagCount++
	if c.__flagCount == 32 {
		binary.LittleEndian.PutUint32(c.__output[c.__flagCursor:], flags)
		c.__flagCount = 0
		c.__flagCursor = c.__outputCursor
		c.__outputCursor += 4
	}
}

func (c *Compressor) __SetEndFlags(flags uint32) {
	if c.__flagCount != 0 {
		flags = (flags << (32 - c.__flagCount)) | ((1 << (32 - c.__flagCount)) - 1)
	} else {
		flags = uint32(0xFFFFFFFF)
	}
	binary.LittleEndian.PutUint32(c.__output[c.__flagCursor:], flags)
}

func (c *Compressor) __literal() {
	c.__output[c.__outputCursor] = c.__input[c.__inputCursor]
	c.__inputCursor++
	c.__outputCursor++
}

func (c *Compressor) __SetUint16(v uint16) {
	binary.LittleEndian.PutUint16(c.__output[c.__outputCursor:], v)
	c.__outputCursor += 2
}

func (c *Compressor) __SetUint32(v uint32) {
	binary.LittleEndian.PutUint32(c.__output[c.__outputCursor:], v)
	c.__outputCursor += 4
}

func (c *Compressor) __SetByte(v uint8) {
	c.__output[c.__outputCursor] = v
	c.__outputCursor++
}

func (c *Compressor) Compress() ([]byte, error) {
	if len(c.__input) == 0 {
		return nil, nil
	}

	c.__output = make([]byte, c.MaxCompressedSize())

	// skip four for flags
	c.__outputCursor += 4
	// copy the first byte
	c.__literal()

	var halfByte *uint8
	flags := uint32(0)
	filledTo := 0

	for c.__inputCursor < len(c.__input)-2 {

		//if c.__outputCursor >= 0x21e892 {
		//	log.Println("debugger break point")
		//}

		if filledTo <= c.__inputCursor {
			filledTo = c.dict.Fill(c.__input, filledTo)
		}
		flags <<= 1
		length, offset := c.dict.Find(c.__input, c.__inputCursor)

		if length < 3 {
			c.__literal()
		} else {
			c.__inputCursor += length
			length -= 3
			c.__SetUint16(uint16(((offset - 1) << 3) | Min(length, 7)))

			if length >= 0x7 {
				length -= 0x7
				if halfByte != nil {
					*halfByte |= byte(Min(length, 0xF) << 4)
					halfByte = nil
				} else {
					halfByte = &c.__output[c.__outputCursor]
					*halfByte = byte(Min(length, 0xF))
					c.__outputCursor++
				}
				if length >= 0xF {
					length -= 0xF
					c.__SetByte(byte(Min(length, 0xFF)))
					if length >= 0xFF {
						length += 0xF + 0x7
						if length <= 0xFFFF {
							c.__SetUint16(uint16(length))
						} else {
							c.__SetUint16(uint16(0))
							c.__SetUint32(uint32(length))
						}
					}
				}
			}
			flags |= 1
		}
		c.__SetFlags(flags)
	}

	for c.__inputCursor < len(c.__input) {
		c.__literal()
		flags <<= 1
		c.__SetFlags(flags)
	}

	c.__SetEndFlags(flags)

	return c.__output[:c.__outputCursor], nil
}

func NewCompressor(input []byte, level Level) *Compressor {
	return &Compressor{
		__input:     input,
		__flagCount: 1,
		dict:        NewDictionary(level),
	}
}

func Min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
