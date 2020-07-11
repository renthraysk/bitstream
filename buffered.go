package bitstream

import "math/bits"

type WriterBuffer struct {
	b   uint64
	out []byte
}

// NewWriterBuffer
func NewWriterBuffer() *WriterBuffer {
	return &WriterBuffer{b: 1, out: make([]byte, 0, 32)}
}

// WriteBit writes a single bit
// x must be 0 or 1; otherwise the behavior is undefined.
func (w *WriterBuffer) WriteBit(x uint64) {
	w.b, x = bits.Add64(w.b, w.b, x)
	if x != 0 {
		w.out = append(w.out, byte(w.b>>56), byte(w.b>>48), byte(w.b>>40), byte(w.b>>32),
			byte(w.b>>24), byte(w.b>>16), byte(w.b>>8), byte(w.b))
		w.b = 1
	}
}

// Len returns number of bits written
func (w *WriterBuffer) Len() int {
	return len(w.out)*8 + bits.Len64(w.b) - 1
}

// Bytes returns the written bits with any 0 bit padding required to meet a byte boundary
func (w *WriterBuffer) Bytes() []byte {
	n := bits.Len64(w.b) - 1 // -1 for sentinel bit
	if n != 0 {
		x := w.b << (64 - n)
		w.b = 1
		w.out = append(w.out, byte(x>>56), byte(x>>48), byte(x>>40), byte(x>>32),
			byte(x>>24), byte(x>>16), byte(x>>8), byte(x))
		w.out = w.out[:len(w.out)-8+(n+7)/8]
	}
	return w.out
}
