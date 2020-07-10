package bitstream

import (
	"encoding/binary"
	"io"
	"math/bits"
)

func boolToUint64(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

// Writer bit writer using a 64 bit buffer
type Writer struct {
	b   uint64
	w   io.Writer
	buf [8]byte
}

// NewWriter returns a Writer that buffers bit writes to the io.Writer w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{b: 1, w: w}
}

// WriteBit writes a single bit
func (w *Writer) WriteBit(b bool) error {
	// :( unable to split this so the 98% path (where c == 0) is able to be inlined.
	x, c := bits.Add64(w.b, w.b, boolToUint64(b))
	if c == 0 {
		w.b = x
		return nil
	}
	binary.BigEndian.PutUint64(w.buf[:8], x)
	_, err := w.w.Write(w.buf[:8])
	w.b = 1
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
	return &Reader{r: r, b: 1 << 63}
}

// ReadBit reads a bit from the underlying reader
func (r *Reader) ReadBit() (bool, error) {
	// :( again splitting this function doesn't give desired inlineability for 96% of calls.
	x, c := bits.Add64(r.b, r.b, 0)
	if c == 0 {
		r.b = x
		return x&(1<<32) != 0, nil
	}
	n, err := r.r.Read(r.buf[:4])
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, io.EOF
	}
	var cc uint32
	y := binary.BigEndian.Uint32(r.buf[:4])
	y, cc = bits.Add32(y, y, 0)
	r.b = 1<<(64-8*n) | uint64(y)
	return cc != 0, nil
}
