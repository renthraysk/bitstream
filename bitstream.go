package bitstream

import (
	"bufio"
	"encoding/binary"
	"io"
	"math/bits"
)

// Writer bit writer using a 64 bit buffer
type Writer struct {
	b   uint64    // 64 bit buffer
	w   io.Writer // destination
	buf [8]byte
}

// NewWriter returns a Writer that buffers bit writes to the io.Writer w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{b: 1, w: w}
}

// WriteBit writes a single bit
// x must be 0 or 1; otherwise the behavior is undefined.
func (w *Writer) WriteBit(x uint64) error {
	// This is written in such away to be inlinable (go 1.14)
	w.b, x = bits.Add64(w.b, w.b, x)
	if x == 0 {
		return nil
	}
	return w.flush()
}

func (w *Writer) flush() error {
	binary.BigEndian.PutUint64(w.buf[:8], w.b)
	w.b = 1
	_, err := w.w.Write(w.buf[:8])
	return err
}

// Flush flushes any bits in the buffer to the output, it pads the output to nearest byte boundary with zero bits.
func (w *Writer) Flush() error {
	n := bits.Len64(w.b) - 1 // -1 for sentinel bit
	if n == 0 {
		return nil
	}
	binary.BigEndian.PutUint64(w.buf[:8], w.b<<(64-n))
	w.b = 1
	_, err := w.w.Write(w.buf[:(n+7)/8])
	return err
}

// Reader reads individual bits from a underlying io.Reader
type Reader struct {
	b   uint64 // This is 2 32 bit values, upper 32 contains a single sentinel bit tracking, lower 32 is our read buffer
	r   io.Reader
	buf [4]byte
}

// NewReader returns a Reader
func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r), b: 1 << 63}
}

// ReadBit reads a bit from the underlying reader
func (r *Reader) ReadBit() (c uint64, err error) {
	r.b, c = bits.Add64(r.b, r.b, 0)
	if c == 0 {
		return (r.b >> 32) & 1, nil
	}
	return r.read()
}

func (r *Reader) read() (uint64, error) {
	n, err := r.r.Read(r.buf[:4])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	var c uint32
	y := binary.BigEndian.Uint32(r.buf[:4])
	y, c = bits.Add32(y, y, 0)
	r.b = 1<<(64-8*n) | uint64(y)
	return uint64(c), nil
}
