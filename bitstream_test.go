package bitstream

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"strconv"
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
		if err := w.WriteBit(1); err != nil {
			b.Fatalf("WriteBit failed: %v", err)
		}
	}
	w.Flush()
}

type Reader55 struct{}

func (Reader55) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 0x55
	}
	return len(b), nil
}

func BenchmarkReadBit(b *testing.B) {
	r := NewReader(&Reader55{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := r.ReadBit(); err != nil {
			b.Fatalf("ReadBit failed: %v", err)
		}
	}
}

func Rand(b *testing.B, n int64) []byte {
	b.Helper()
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		b.Fatalf("failed to read rand: %v", err)
	}
	return buf
}

func BenchCopy(b *testing.B, n int64) {
	b.Helper()

	ww := &bytes.Buffer{}
	rr := bytes.NewReader(Rand(b, n))

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(n)

	for i := 0; i < b.N; i++ {
		ww.Reset()
		rr.Seek(0, io.SeekStart)

		r := NewReader(rr)
		w := NewWriter(ww)

		for b, err := r.ReadBit(); err == nil; b, err = r.ReadBit() {
			w.WriteBit(b)
		}
		w.Flush()
	}
}

func BenchmarkRate256(b *testing.B) {
	BenchCopy(b, 256)
}

func BenchmarkRate1K(b *testing.B) {
	BenchCopy(b, 1024)
}

func BenchmarkRate4K(b *testing.B) {
	BenchCopy(b, 4096)
}

func BenchmarkRate10K(b *testing.B) {
	BenchCopy(b, 10240)
}

func TestWriteBits(t *testing.T) {

	type wb struct {
		x uint16
		n int
	}

	tests := []struct {
		pre      wb
		bytes    []byte
		post     wb
		expected []byte
	}{
		0:  {pre: wb{x: 0b0, n: 1}, expected: []byte{0x00}},
		1:  {pre: wb{x: 0b1, n: 1}, expected: []byte{0x80}},
		2:  {pre: wb{x: 0b11, n: 2}, expected: []byte{0xC0}},
		3:  {pre: wb{x: 0b111, n: 3}, expected: []byte{0xE0}},
		4:  {pre: wb{x: 0b1111, n: 4}, expected: []byte{0xF0}},
		5:  {pre: wb{x: 0b1111111, n: 7}, expected: []byte{0xFE}},
		6:  {pre: wb{x: 0b11111111, n: 8}, expected: []byte{0xFF}},
		7:  {pre: wb{x: 0b111111111, n: 9}, expected: []byte{0xFF, 0x80}},
		8:  {pre: wb{x: ^uint16(0), n: 16}, expected: []byte{0xFF, 0xFF}},
		9:  {pre: wb{x: 0, n: 1}, bytes: []byte{0xFF}, expected: []byte{0x7F, 0x80}},
		10: {pre: wb{x: 0, n: 1}, bytes: []byte{0xFF, 0x00}, expected: []byte{0x7F, 0x80, 0x00}},
		11: {pre: wb{x: 0, n: 1}, bytes: []byte{0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00}, expected: []byte{0x7F, 0x80, 0x7F, 0x80, 0x7F, 0x80, 0x7F, 0x80, 0x00}},
		12: {pre: wb{x: 0, n: 1}, bytes: []byte{0xFF}, expected: []byte{0x7F, 0x80}},

		13: {bytes: []byte{0, 0xFF, 0, 0xFF, 0, 0xFF, 0x00}, post: wb{x: 0xF, n: 4}, expected: []byte{0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xF0}},
	}
	for i, ts := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ww := &bytes.Buffer{}
			w := NewWriter(ww)

			if ts.pre.n > 0 {
				w.WriteBits(ts.pre.x, ts.pre.n)
			}
			for _, x := range ts.bytes {
				w.WriteByte(x)
			}
			if ts.post.n > 0 {
				w.WriteBits(ts.post.x, ts.post.n)
			}
			w.Flush()

			if !bytes.Equal(ww.Bytes(), ts.expected) {
				t.Fatalf("failure expected %x, got %x", ts.expected, ww.Bytes())

			}
		})
	}
}
