# bitstream

Test of an idea of using bits.Add() functions for a fast bit reading and writing from io.Reader and io.Writers respectively. 
The idea was to use a sentinel like bit in both reading and writing which when returned in bits.Add() carry, triggers the actual reading and writing to underlying io.Reader or io.Writer.
