package govarint

import "bytes"
import "io"
import "testing"

var fourU32 = []uint32{
	0,
	1,
	0,
	256,
}

var fiveU32 = []uint32{
	42,
	4294967196,
	384,
	9716053,
	1024 + 256 + 3,
}

var testU32 = []uint32{
	0,
	1,
	2,
	10,
	20,
	63,
	64,
	65,
	127,
	128,
	129,
	255,
	256,
	257,
}

var testU64 = []uint64{
	0,
	1,
	2,
	10,
	20,
	63,
	64,
	65,
	127,
	128,
	129,
	255,
	256,
	257,
	///
	1<<32 - 1,
	1 << 32,
	1 << 33,
	1 << 42,
	1<<63 - 1,
	1 << 63,
}

func TestEncodeAndDecodeU32(t *testing.T) {
	for _, expected := range testU32 {
		var buf bytes.Buffer
		enc := NewU32Base128Encoder(&buf)
		enc.PutU32(expected)
		enc.Close()
		dec := NewU32Base128Decoder(&buf)
		x, err := dec.GetU32()
		if x != expected || err != nil {
			t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, expected, err)
		}
	}
	var buf bytes.Buffer
	enc := NewU32Base128Encoder(&buf)
	for _, expected := range testU32 {
		enc.PutU32(expected)
	}
	enc.Close()
	dec := NewU32Base128Decoder(&buf)
	i := 0
	for {
		x, err := dec.GetU32()
		if err == io.EOF {
			break
		}
		if x != testU32[i] || err != nil {
			t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, testU32[i], err)
		}
		i += 1
	}
	if i != len(testU32) {
		t.Errorf("Only %d integers were decoded when %d were encoded", i, len(testU32))
	}
}

func TestEncodeAndDecodeU64(t *testing.T) {
	for _, expected := range testU64 {
		var buf bytes.Buffer
		enc := NewU64Base128Encoder(&buf)
		enc.PutU64(expected)
		enc.Close()
		dec := NewU64Base128Decoder(&buf)
		x, err := dec.GetU64()
		if x != expected || err != nil {
			t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, expected, err)
		}
	}
}

func TestU32GroupVarintFour(t *testing.T) {
	var buf bytes.Buffer
	enc := NewU32GroupVarintEncoder(&buf)
	for _, expected := range fourU32 {
		enc.PutU32(expected)
	}
	enc.Close()
	dec := NewU32GroupVarintDecoder(&buf)
	i := 0
	for {
		x, err := dec.GetU32()
		if err == io.EOF {
			break
		}
		if err != nil && x != fourU32[i] {
			t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, testU32[i], err)
		}
		i += 1
	}
	if i != len(fourU32) {
		t.Errorf("%d integers were decoded when %d were encoded", i, len(fourU32))
	}
}

func TestU32GroupVarintFive(t *testing.T) {
	var buf bytes.Buffer
	enc := NewU32GroupVarintEncoder(&buf)
	for _, expected := range fiveU32 {
		enc.PutU32(expected)
	}
	enc.Close()
	dec := NewU32GroupVarintDecoder(&buf)
	i := 0
	for {
		x, err := dec.GetU32()
		if err == io.EOF {
			break
		}
		if err != nil && x != fiveU32[i] {
			t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, testU32[i], err)
		}
		i += 1
	}
	if i != len(fiveU32) {
		t.Errorf("%d integers were decoded when %d were encoded", i, len(fiveU32))
	}
}

func TestU32GroupVarint14(t *testing.T) {
	var buf bytes.Buffer
	for length := 0; length < len(testU32); length++ {
		subset := testU32[:length]
		enc := NewU32GroupVarintEncoder(&buf)
		for _, expected := range subset {
			enc.PutU32(expected)
		}
		enc.Close()
		dec := NewU32GroupVarintDecoder(&buf)
		i := 0
		for {
			x, err := dec.GetU32()
			if err == io.EOF {
				break
			}
			if err != nil && x != subset[i] {
				t.Errorf("ReadUvarint(%v): got x = %d, expected = %d, err = %s", buf, x, subset[i], err)
			}
			i += 1
		}
		if i != len(subset) {
			t.Errorf("%d integers were decoded when %d were encoded", i, len(subset))
		}
	}
}
