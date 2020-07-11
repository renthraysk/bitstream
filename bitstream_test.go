package bitstream

import (
	"crypto/rand"
	"io/ioutil"
	"testing"

	"bytes"
)

func TestCopy(t *testing.T) {

	tests := [][]byte{
		{0x00},
		{0xFF},
		{0x55},
		{0xAA},
		{0x00, 0x01, 0x02},
		{0x00, 0x01, 0x02, 0x03},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
	}

	for _, ts := range tests {
		in := bytes.NewBuffer(ts)
		var out bytes.Buffer

		r := NewReader(in)
		w := NewWriter(&out)

		for b, err := r.ReadBit(); err == nil; b, err = r.ReadBit() {
			w.WriteBit(b)
		}
		w.Flush()

		if !bytes.Equal(ts, out.Bytes()) {
			t.Fatalf("equal failed: expected %x got %x", ts, out.Bytes())
		}
	}
}

func BenchmarkWriteBit(b *testing.B) {
	w := NewWriter(ioutil.Discard)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := w.WriteBit(uint64(i) & 1); err != nil {
			b.Fatalf("WriteBit failed: %v", err)
		}
	}
	w.Flush()
}

func BenchmarkBufferWriteBit(b *testing.B) {
	w := NewWriterBuffer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.WriteBit(uint64(i) & 1)
	}
	w.Bytes()
}

func BenchmarkReadBit(b *testing.B) {
	r := NewReader(rand.Reader)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := r.ReadBit(); err != nil {
			b.Fatalf("ReadBit failed: %v", err)
		}
	}
}
