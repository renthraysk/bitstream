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
	b   uint64 // Upto 63 buffered bits, and one sentinel bit
	r   io.Reader
	buf [8]byte
}

// NewReader returns a Reader
func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

// ReadBit reads a bit from the underlying reader
func (r *Reader) ReadBit() (c uint64, err error) {
	r.b, c = bits.Add64(r.b, r.b, 0)
	if r.b == 0 {
		// Sentinel bit isn't present, need to fill buffer
		return r.fill()
	}
	return
}

func (r *Reader) fill() (c uint64, err error) {
	// clear buffer incase of short reads
	binary.LittleEndian.PutUint64(r.buf[:8], 0)
	n, err := r.r.Read(r.buf[:8])
	if err != nil {
		return
	}
	if n == 0 {
		return 0, io.EOF
	}
	y := binary.BigEndian.Uint64(r.buf[:8])
	// Pop a bit off freeing a bit for the sentinel
	y, c = bits.Add64(y, y, 0)
	r.b = y | 1<<(64-8*n)
	return
}
