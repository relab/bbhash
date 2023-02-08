package bbhash

import (
	"bytes"
	"math/bits"
)

// bitVector represents a bit vector in an efficient manner.
type bitVector struct {
	v []uint64
}

// newBitVector creates a bit vector with the given number of words.
func newBitVector(words uint64) *bitVector {
	return &bitVector{
		v: make([]uint64, words),
	}
}

// words returns the number of words the bit vector needs to hold size bits, with expansion factor gamma.
func words(size uint64, gamma float64) uint64 {
	sz := uint64(float64(size) * gamma)
	return (sz + 63) / 64
}

// Size returns the number of bits this bit vector has allocated.
func (b *bitVector) Size() uint64 {
	return uint64(len(b.v) * 64)
}

// Words returns the number of 64-bit words this bit vector has allocated.
func (b *bitVector) Words() uint64 {
	return uint64(len(b.v))
}

// Set sets the bit at position i.
func (b *bitVector) Set(i uint64) {
	b.v[i/64] |= 1 << (i % 64)
}

// Unset zeroes the bit at position i.
func (b *bitVector) Unset(i uint64) {
	b.v[i/64] &^= 1 << (i % 64)
}

// IsSet returns true if the bit at position i is set.
func (b *bitVector) IsSet(i uint64) bool {
	return b.v[i/64]&(1<<(i%64)) != 0
}

// Reset reduces the bit vector's size to words and zeroes the elements.
func (b *bitVector) Reset(words uint64) {
	b.v = b.v[:words]
	for i := range b.v {
		b.v[i] = 0
	}
}

// OnesCount returns the number of one bits ("population count") in the bit vector.
func (b *bitVector) OnesCount() uint64 {
	var p int
	for i := range b.v {
		p += bits.OnesCount64(b.v[i])
	}
	return uint64(p)
}

// Rank returns the number of one bits in the bit vector up to position i.
func (b *bitVector) Rank(i uint64) uint64 {
	x := i / 64
	y := i % 64

	var r int
	for k := uint64(0); k < x; k++ {
		r += bits.OnesCount64(b.v[k])
	}
	v := b.v[x]
	r += bits.OnesCount64(v << (64 - y))
	return uint64(r)
}

// String returns a string representation of the bit vector.
func (b *bitVector) String() string {
	var buf bytes.Buffer
	for i := uint64(0); i < b.Size(); i++ {
		if b.IsSet(i) {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
	}
	return buf.String()
}
