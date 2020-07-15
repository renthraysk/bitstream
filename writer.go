package bitstream

import (
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
	if x != 0 {
		// Sentinel bit carried off, w.b is full
		return w.flushAll()
	}
	return nil
}

func (w *Writer) flushAll() error {
	binary.BigEndian.PutUint64(w.buf[:8], w.b)
	w.b = 1
	_, err := w.w.Write(w.buf[:8])
	return err
}

// WriteByte writes a single byte
func (w *Writer) WriteByte(b byte) (err error) {
	// As the sentinel bit is highest bit set
	// can use a simple compare to see if have enough empty bits available
	if w.b >= 1<<56 {
		err = w.flush()
	}
	w.b = w.b<<8 | uint64(b)
	return
}

// WriteBits writes the lowest n bits of x, with the most significant bit of the lower n bits first.
// n must be between 1 and 16.
func (w *Writer) WriteBits(x uint16, n int) (err error) {
	if w.b >= 1<<48 {
		err = w.flush()
	}
	w.b = w.b<<n | uint64(x)
	return
}

func (w *Writer) flush() error {
	n := bits.Len64(w.b) - 1
	binary.BigEndian.PutUint64(w.buf[:8], w.b<<(64-n))
	s := uint64(1) << (n & 7)
	w.b = s | (w.b & (s - 1))
	_, err := w.w.Write(w.buf[:n>>3])
	return err
}

// Flush flushes any bits in the buffer to the output, it pads the output to nearest byte boundary with zero bits.
func (w *Writer) Flush() error {
	n := bits.Len64(w.b) // n includes sentinel bit
	if n <= 1 {
		return nil
	}
	binary.BigEndian.PutUint64(w.buf[:8], w.b<<(65-n))
	w.b = 1
	_, err := w.w.Write(w.buf[:(n+6)/8])
	return err
}
