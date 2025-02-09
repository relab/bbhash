package bbhash

import (
	"math/bits"
	"strconv"
	"strings"
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
func words(size int, gamma float64) uint64 {
	sz := uint64(float64(size) * gamma)
	return (sz + 63) / 64
}

// size returns the number of bits this bit vector has allocated.
func (b *bitVector) size() uint64 {
	return uint64(len(b.v) * 64)
}

// words returns the number of 64-bit words this bit vector has allocated.
func (b *bitVector) words() uint64 {
	return uint64(len(b.v))
}

// set sets the bit at position i.
func (b *bitVector) set(i uint64) {
	b.v[i/64] |= 1 << (i % 64)
}

// unset zeroes the bit at position i.
func (b *bitVector) unset(i uint64) {
	b.v[i/64] &^= 1 << (i % 64)
}

// isSet returns true if the bit at position i is set.
func (b *bitVector) isSet(i uint64) bool {
	return b.v[i/64]&(1<<(i%64)) != 0
}

// reset reduces the bit vector's size to words and zeroes the elements.
func (b *bitVector) reset(words uint64) {
	b.v = b.v[:words]
	clear(b.v)
}

// onesCount returns the number of one bits ("population count") in the bit vector.
func (b *bitVector) onesCount() uint64 {
	var p int
	for i := range b.v {
		p += bits.OnesCount64(b.v[i])
	}
	return uint64(p)
}

// rank returns the number of one bits in the bit vector up to position i.
func (b *bitVector) rank(i uint64) uint64 {
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
	var buf strings.Builder
	for i := uint64(0); i < b.size(); i++ {
		if b.isSet(i) {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
	}
	return buf.String()
}

// stringList returns a string list of true positions in the bit vector.
// Mainly useful for debugging with smaller bit vectors.
func (b *bitVector) stringList() string {
	var buf strings.Builder
	buf.WriteString("(")
	for i := uint64(0); i < b.size(); i++ {
		if b.isSet(i) {
			buf.WriteString(strconv.Itoa(int(i)))
			if i < b.size()-1 {
				buf.WriteString(", ")
			}
		}
	}
	buf.WriteString(")")
	return buf.String()
}
