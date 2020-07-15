package bitstream

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"math/bits"
)

// Reader reads individual bits from a underlying io.Reader
type Reader struct {
	b   uint64 // Upto 63 buffered bits, and one set sentinel bit
	r   io.Reader
	buf [8]byte
}

// NewReader returns a Reader
func NewReader(r io.Reader) *Reader {
	if _, ok := r.(*bytes.Reader); ok {
		return &Reader{r: r}
	}
	return &Reader{r: bufio.NewReader(r)}
}

// ReadBit reads a bit from the underlying reader
func (r *Reader) ReadBit() (b uint64, err error) {
	r.b, b = bits.Add64(r.b, r.b, 0)
	if r.b == 0 {
		// Sentinel bit is no longer present, need to fill buffer
		return r.fill()
	}
	return
}

func (r *Reader) fill() (b uint64, err error) {
	// clear buffer incase of short reads
	binary.LittleEndian.PutUint64(r.buf[:8], 0)
	n, err := r.r.Read(r.buf[:8])
	if err != nil {
		return
	}
	if n == 0 {
		return 0, io.EOF
	}
	x := binary.BigEndian.Uint64(r.buf[:8])
	// Pop a bit off ensuring a free bit for the sentinel
	x, b = bits.Add64(x, x, 0)
	r.b = x | 1<<(64-8*n)
	return
}
