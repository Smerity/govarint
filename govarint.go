package govarint

import "encoding/binary"
import "io"

type U32VarintEncoder interface {
	GetU32(r io.ByteReader) (uint32, error)
	PutU32(x uint32) int
	Close()
}

type U32VarintDecoder interface {
	GetU32() (uint32, error)
}

///

type U64VarintEncoder interface {
	GetU64(r io.ByteReader) (uint64, error)
	PutU64(x uint64) int
	Close()
}

type U64VarintDecoder interface {
	GetU64() (uint64, error)
}

///

type U32GroupVarintEncoder struct {
	w     io.Writer
	index int
	store [4]uint32
	temp  [17]byte
}

func NewU32GroupVarintEncoder(w io.Writer) *U32GroupVarintEncoder { return &U32GroupVarintEncoder{w: w} }

func (b *U32GroupVarintEncoder) Flush() int {
	// TODO: Is it more efficient to have a tailored version that's called only in Close()?
	if b.index == 0 {
		return 0
	}
	for i := b.index; i < 4; i++ {
		b.store[i] = 0
	}
	length := 1
	for i, x := range b.store {
		size := byte(0)
		shifts := []byte{24, 16, 8, 0}
		for _, shift := range shifts {
			if shift == 0 || (x>>shift) != 0 {
				size += 1
				b.temp[length] = byte(x >> shift)
				length += 1
			}
		}
		size -= 1
		b.temp[0] |= size << (uint8(3-i) * 2)
	}
	if b.index != 4 {
		length -= 4 - b.index
	}
	b.w.Write(b.temp[:length])
	return length
}

func (b *U32GroupVarintEncoder) PutU32(x uint32) (int, error) {
	bytesWritten := 0
	b.store[b.index] = x
	b.index += 1
	if b.index == 4 {
		bytesWritten += b.Flush()
		b.index = 0
	}
	return bytesWritten, nil
}

func (b *U32GroupVarintEncoder) Close() {
	b.Flush()
}

///

type U32GroupVarintDecoder struct {
	r        io.ByteReader
	group    [4]uint32
	pos      int
	finished bool
	capacity int
}

func NewU32GroupVarintDecoder(r io.ByteReader) *U32GroupVarintDecoder {
	return &U32GroupVarintDecoder{r: r, pos: 4, capacity: 4}
}

func (b *U32GroupVarintDecoder) getGroup() error {
	sizeByte, err := b.r.ReadByte()
	if err != nil {
		return err
	}
	for index := 0; index < 4; index++ {
		b.group[index] = 0
		size := int(sizeByte>>(uint8(3-index)*2))&0x03 + 1
		for i := 0; i < size; i++ {
			valByte, err := b.r.ReadByte()
			if err == io.EOF {
				b.capacity = index
				b.finished = true
			} else if err != nil {
				return err
			}
			b.group[index] = b.group[index]<<8 + uint32(valByte)
		}
		if b.finished {
			break
		}
	}
	b.pos = 0
	return nil
}

func (b *U32GroupVarintDecoder) GetU32() (uint32, error) {
	if b.pos == b.capacity {
		if b.finished {
			return 0, io.EOF
		}
		err := b.getGroup()
		if err != nil {
			return 0, err
		}
	}
	b.pos += 1
	return b.group[b.pos-1], nil
}

///

type Base128Encoder struct {
	w io.Writer
}

func NewU32Base128Encoder(w io.Writer) *Base128Encoder { return &Base128Encoder{w: w} }
func NewU64Base128Encoder(w io.Writer) *Base128Encoder { return &Base128Encoder{w: w} }

func (b *Base128Encoder) PutU32(x uint32) (int, error) {
	bytes := make([]byte, binary.MaxVarintLen32)
	writtenBytes := binary.PutUvarint(bytes, uint64(x))
	return b.w.Write(bytes[:writtenBytes])
}

func (b *Base128Encoder) PutU64(x uint64) (int, error) {
	bytes := make([]byte, binary.MaxVarintLen64)
	writtenBytes := binary.PutUvarint(bytes, x)
	return b.w.Write(bytes[:writtenBytes])
}

func (b *Base128Encoder) Close() {
}

///

type Base128Decoder struct {
	r io.ByteReader
}

func NewU32Base128Decoder(r io.ByteReader) *Base128Decoder { return &Base128Decoder{r: r} }
func NewU64Base128Decoder(r io.ByteReader) *Base128Decoder { return &Base128Decoder{r: r} }

func (b *Base128Decoder) GetU32() (uint32, error) {
	v, err := binary.ReadUvarint(b.r)
	return uint32(v), err
}

func (b *Base128Decoder) GetU64() (uint64, error) {
	return binary.ReadUvarint(b.r)
}
