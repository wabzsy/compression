package aplib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/bits"
	"unsafe"
)

type BitCompressor struct {
	__tagSize   uint8
	__bitBuffer uint8
	__tagOffset int
	__maxBit    int
	__bitCount  int
	__isTagged  bool

	buffer *bytes.Buffer
}

func NewBitCompressor() *BitCompressor {
	return &BitCompressor{
		buffer:    &bytes.Buffer{},
		__tagSize: 1,
		__maxBit:  7, // tagSize * 8 -1
	}
}

func (c *BitCompressor) writeBitSequence(bits ...int) {
	for _, bit := range bits {
		c.writeBit(bit)
	}
}

func (c *BitCompressor) writeFixedNumber(value, nBits int) {
	for i := nBits - 1; i >= 0; i-- {
		c.writeBit((value >> i) & 1)
	}
}

func (c *BitCompressor) writeVariableNumber(value int) {
	if value < 2 {
		panic(fmt.Sprintf("writeVariableNumber: value=%d", value))
	}

	length := bits.Len(uint(value)) - 2

	c.writeBit(value & (1 << length))

	for i := length - 1; i >= 0; i-- {
		c.writeBit(1)
		c.writeBit(value & (1 << i))
	}

	c.writeBit(0)
}

func LengthDelta(offset int) int {
	if offset < 0x80 || 0x7D00 <= offset {
		return 2
	} else if 0x500 <= offset {
		return 1
	}
	return 0
}
func (c *BitCompressor) updateTag(end bool) {
	// tagOffset == 0 说明还没打标签,无需写入
	if c.__tagOffset != 0 {
		c.buffer.Bytes()[c.__tagOffset] = c.__bitBuffer
	}

	// 如果不是最后一次更新, 写入下一个tag的占位符
	if !end {
		// 移动tagOffset到缓冲区结尾
		c.__tagOffset = c.buffer.Len()
		// 写入tag占位符, tagSize固定为1
		c.writeByte(c.__tagSize)
	}
}

func (c *BitCompressor) writeBit(bit int) {
	// bitCount为0说明已写满1字节或第一次写
	if c.__bitCount == 0 {
		// 当已满1字节时更新标签
		c.updateTag(false)
		// 重置bit记录器
		c.__bitCount = c.__maxBit
		c.__bitBuffer = 0
	} else {
		// 不等于0说明当前字节还没写满
		c.__bitCount--
	}

	if bit != 0 { // 仅在存在bit位时才进行运算,
		c.__bitBuffer |= 1 << c.__bitCount
	}
}

func (c *BitCompressor) writeByte(n uint8) {
	c.buffer.WriteByte(n)
}

func (c *BitCompressor) Bytes() []byte {
	return c.buffer.Bytes()
}

type Compressor struct {
	*BitCompressor

	__input       []byte
	__inputCursor int
	__lastOffset  int
	__pair        bool
}

func NewCompressor(input []byte) *Compressor {
	return &Compressor{
		BitCompressor: NewBitCompressor(),
		__input:       input,
		__pair:        true,
	}
}

/*
false: buffer 添加当前光标的字节, 光标+1 标记为已配对
true: 写入一个bit0(不知道是否是用于gamma判断结束的那个),  buffer 添加当前光标的字节, 光标+1 标记为已配对
*/

func (c *Compressor) __literal(marker bool) {
	if marker {
		c.writeBit(0)
	}

	c.writeByte(c.__input[c.__inputCursor])
	c.__inputCursor += 1
	c.__pair = true
}

func (c *Compressor) __singleByte(offset int) {
	if !(0 <= offset && offset < 16) {
		panic(fmt.Sprintf("__singleByte: offset=%d", offset))
	}

	c.writeBitSequence(1, 1, 1)
	c.writeFixedNumber(offset, 4)
	c.__inputCursor += 1
	c.__pair = true
}

func (c *Compressor) __shortBlock(offset, length int) {
	if !(0 < offset && offset <= 127) ||
		!(2 <= length && length <= 3) {
		panic(fmt.Sprintf("__shortBlock: offset=%d, length=%d", offset, length))
	}

	c.writeBitSequence(1, 1, 0) // short block
	// offset 最大 0b01111111 最小 0b00000001, length 只有两种可能: 2 或 3, 如果是2:-2=0, 如果是3:-2=1
	// 将-2之后的length拼接到offset<<1后的最低位, 变成0b11111110+0或0b11111110+1, 所以这里理论上还可以用|运算符代替+运算符
	c.writeByte(byte((offset << 1) + (length - 2)))
	c.__inputCursor += length
	c.__lastOffset = offset
	c.__pair = false
}

func (c *Compressor) __block(offset, length int) {
	if !(offset >= 2) {
		panic(fmt.Sprintf("__block: offset=%d, length=%d", offset, length))
	}

	c.writeBitSequence(1, 0)

	if c.__pair && c.__lastOffset == offset { // 重用offset
		c.writeVariableNumber(2)
		c.writeVariableNumber(length)
	} else {
		high := (offset >> 8) + 2
		if c.__pair {
			high += 1
		}
		c.writeVariableNumber(high)
		c.writeByte(byte(offset & 0xFF))
		c.writeVariableNumber(length - LengthDelta(offset))
	}

	c.__inputCursor += length
	c.__lastOffset = offset
	c.__pair = false
}

func (c *Compressor) __end() {
	c.writeBitSequence(1, 1, 0)
	c.writeByte(0)
	c.updateTag(true)
}

func (c *Compressor) Compress(safe bool) ([]byte, error) {
	result := c.Pack()

	if safe {
		header := AP32Header{
			Magic:      [4]byte{'A', 'P', '3', '2'},
			HeaderSize: 24,
			PackedSize: uint32(len(result)),
			PackedCrc:  crc32.ChecksumIEEE(result),
			OrigSize:   uint32(len(c.__input)),
			OrigCrc:    crc32.ChecksumIEEE(c.__input),
		}

		header.HeaderSize = uint32(unsafe.Sizeof(header))

		headerBuf := bytes.NewBuffer(nil)
		if err := binary.Write(headerBuf, binary.LittleEndian, &header); err != nil {
			return nil, err
		}

		result = append(headerBuf.Bytes(), result...)
	}

	return result, nil
}

func (c *Compressor) Pack() []byte {
	c.__literal(false)
	for c.__inputCursor < len(c.__input) {
		offset, length := Search(c.__input, c.__inputCursor) // 当前[cursor:cursor+length]==之前[cursor-offset:cursor-offset+length]

		// fmt.Println("cursor:", c.__inputCursor, "\t", offset, length)
		if length == 0 { // 未找到子串时
			if c.__input[c.__inputCursor] == 0 { // 判断当前字符是否为0x00
				c.__singleByte(0)
			} else {
				c.__literal(true)
			}
		} else if length == 1 && 0 <= offset && offset < 16 {
			c.__singleByte(offset) // 1 1 1
		} else if 2 <= length && length <= 3 && 0 < offset && offset <= 127 {
			c.__shortBlock(offset, length) // 1 1 0
		} else if 3 < length && 2 <= offset {
			c.__block(offset, length) // 1 0
		} else {
			c.__literal(true) // 0
		}
	}

	c.__end()
	return c.Bytes()
}

const maxWindowSize = 8 * 1024

func Search(buf []byte, cursor int) (offset, length int) { // py2改

	//begin := time.Now()
	//defer func() {
	//	fmt.Println("cursor:", cursor, "time:", time.Since(begin), "offset:", offset, "length:", length)
	//}()

	start := cursor - maxWindowSize
	if start < 0 {
		start = 0
	}

	search := buf[start:cursor]
	ahead := buf[cursor:]

	for idx := 0; idx < len(ahead); idx++ {
		// word := ahead[:idx+1]
		if position := LastIndexOf(search, ahead[:idx+1]); position == -1 {
			return
		} else {
			offset = cursor - start - position
			length = idx + 1
		}
	}
	return
}

func LastIndexOf(search, word []byte) int {
	searchLen := len(search)
	wordLen := len(word)
	byteOffset := searchLen - 1

	if byteOffset+wordLen > searchLen {
		byteOffset = searchLen - wordLen
	}

	for i := byteOffset; i >= 0; i-- {
		found := true
		for j := 0; j < wordLen; j++ {
			if search[i+j] != word[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}

	return -1
}
