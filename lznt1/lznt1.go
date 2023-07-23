package lznt1

import (
	"bytes"
	"encoding/binary"
	"math"
)

const (
	CHUNK_SIZE = 0x1000
)

type Entry struct {
	Pos map[int16]int
}

func NewEntry() *Entry {
	return &Entry{
		Pos: make(map[int16]int),
	}
}

type Directory struct {
	Entries map[uint16]*Entry
	Cursor  int
	Sizes   []int16
}

func (d *Directory) LoadEntry(idx uint16) *Entry {
	// TODO: 考虑未来是否会出现并发map的情况
	if d.Entries[idx] == nil {
		d.Entries[idx] = NewEntry()
	}
	return d.Entries[idx]
}

func (d *Directory) Fill(input []byte, cursor int, length int) {
	d.Entries = make(map[uint16]*Entry)
	d.Sizes = make([]int16, 0x100*0x100)
	d.Cursor = cursor

	// 添加条目
	for i := 0; i < length-2; i++ {
		idx := uint16(input[cursor+i])<<8 | uint16(input[cursor+i+1])
		d.LoadEntry(idx).Pos[d.Sizes[idx]] = cursor + i
		d.Sizes[idx]++
	}
}

func (d *Directory) Find(input []byte, cursor int, maxLen int) (offset, length int) {
	if maxLen < 3 || cursor <= d.Cursor {
		return 0, 0
	}

	idx := uint16(input[cursor])<<8 | uint16(input[cursor+1])
	entryPos := d.LoadEntry(idx).Pos

	// Do an exhaustive search (with the possible positions)
	for j := int16(0); j < d.Sizes[idx]-1 && entryPos[j] < cursor; j++ {
		position := entryPos[j]
		i := 2

		// TODO: 此处还可继续优化
		for i < maxLen && input[cursor+i] == input[position+i] {
			i++
		}

		if i > length {
			offset = cursor - position
			length = i
			if length == maxLen {
				break
			}
		}
	}

	if length >= 3 {
		return offset, length
	}

	return 0, 0
}

func NewDirectory() *Directory {
	return &Directory{
		Entries: make(map[uint16]*Entry),
		Sizes:   make([]int16, 0x100*0x100),
	}
}

type Compressor struct {
	dir            *Directory
	__input        []byte
	__inputCursor  int
	__output       []byte
	__outputCursor int
	buffer         *bytes.Buffer
}

func (c *Compressor) compressChunk(chunkLength int) int {
	// 初始化Directory
	c.dir.Fill(c.__input, c.__inputCursor, chunkLength)

	out := c.__output[c.__outputCursor+2:]

	inPos, outPos := 0, 0
	rem := chunkLength
	pow2 := 0x10
	mask3 := 0x1002
	shift := 12

	for rem != 0 {
		i := 0
		pos := 0
		bits := 0
		var bs [16]byte

		// Go through each bit
		// if all are special, then it will fill 16 bytes
		for ; i < 8 && rem != 0; i++ {
			bits >>= 1
			for pow2 < inPos {
				pow2 <<= 1
				mask3 = (mask3 >> 1) + 1
				shift--
			}

			offset, length := c.dir.Find(c.__input, c.__inputCursor+inPos, Min(rem, mask3))

			if length > 0 {
				// Write symbol that is a combination of offset and length
				sym := uint16(((offset - 1) << shift) | (length - 3))
				binary.LittleEndian.PutUint16(bs[pos:], sym)
				pos += 2
				bits |= 0x80 // set the highest bit
				inPos += length
				rem -= length
			} else {
				// 未找到匹配写当前字节
				bs[pos] = c.__input[c.__inputCursor+inPos]
				pos++
				inPos++
				rem--
			}
		}

		if outPos+1+pos >= chunkLength {
			// 说明chunk无法被压缩,或压缩后体积更大
			return chunkLength
		}

		out[outPos] = byte(bits >> (8 - i)) // finish moving the value over
		copy(out[outPos+1:], bs[:pos])
		outPos += 1 + pos
	}

	if rem == 0 {
		return outPos
	} else {
		// 说明chunk无法被压缩,或压缩后体积更大
		return chunkLength
	}
}

func (c *Compressor) MaxCompressedSize() int {
	return len(c.__input) + 3 + 2*((len(c.__input)+CHUNK_SIZE-1)/CHUNK_SIZE)
}

func (c *Compressor) Compress() ([]byte, error) {
	c.__output = make([]byte, c.MaxCompressedSize())

	for c.__inputCursor < len(c.__input) {
		// Compress the next chunk
		chunkLength := Min(len(c.__input)-c.__inputCursor, CHUNK_SIZE)
		compressedLength := c.compressChunk(chunkLength)

		flags := 0
		if compressedLength < chunkLength {
			// 压缩成功, 数据已在compressChunk里写入了
			flags = 0xB000
		} else {
			// chunk is uncompressed
			compressedLength = chunkLength
			flags = 0x3000
			// 跳过header的位置 先写后面的数据, header在下面统一写
			copy(c.__output[c.__outputCursor+2:], c.__input[c.__inputCursor:c.__inputCursor+compressedLength])
		}

		// 写入header(2 byte)
		binary.LittleEndian.PutUint16(c.__output[c.__outputCursor:], uint16(flags|(compressedLength-1)))
		// 输出位置后移: 本轮压缩后的大小+2(header)
		c.__outputCursor += compressedLength + 2
		// 输入位置后移: 本轮压缩的大小
		c.__inputCursor += chunkLength
	}

	return c.__output[:c.__outputCursor], nil
}

func NewCompressor(input []byte) *Compressor {
	return &Compressor{
		dir:     NewDirectory(),
		__input: input,
		buffer:  &bytes.Buffer{},
	}
}

func Min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func Compress(input []byte) ([]byte, error) {
	return NewCompressor(input).Compress()
}
