package bitstream

import (
	"crypto/rand"
	"io"
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
