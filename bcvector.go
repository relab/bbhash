package bbhash

// bcVector represents a combined bit and collision vector.
type bcVector struct {
	v []uint64
	c []uint64
}

// newBCVector returns a combined bit and collision vector with the given number of words.
func newBCVector(words uint64) *bcVector {
	return &bcVector{
		v: make([]uint64, words),
		c: make([]uint64, words),
	}
}

// nextLevel returns a new combined bit and collision vector with the given number of words.
// The collision vector is reused from the original combined vector.
func (b *bcVector) nextLevel(words uint64) {
	b.c = b.c[:words]
	for i := range b.c {
		b.c[i] = 0
	}
	b.v = make([]uint64, words)
}

func (b *bcVector) bitVector() *bitVector {
	return &bitVector{v: b.v}
}

func (b *bcVector) collisionVector() *bitVector {
	return &bitVector{v: b.c}
}

// Size returns the number of bits this bit vector has allocated.
func (b *bcVector) Size() uint64 {
	return uint64(len(b.v) * 64)
}

// Update sets the bit at position i, and records a collision if the bit was already set.
func (b *bcVector) Update(i uint64) {
	x, y := i/64, uint64(1<<(i%64))
	if b.v[x]&y != 0 {
		// found one or more collisions at index i; update collision vector
		b.c[x] |= y
		return
	}
	// no collisions at index i; set bit
	b.v[x] |= y
}

// UnsetCollision returns true if the bit at position i has a collision.
// The vector is also unset at position i.
func (b *bcVector) UnsetCollision(i uint64) bool {
	x, y := i/64, uint64(1<<(i%64))
	if b.c[x]&y != 0 {
		// found collision at index i; unset bit
		b.v[x] &^= y
		return true
	}
	// no collisions at index i
	return false
}

// String returns a string representation of the bit vector.
func (b *bcVector) String() string {
	return b.bitVector().String()
}

// Collisions returns a string representation of the collision vector.
func (b *bcVector) Collisions() string {
	return b.collisionVector().String()
}

// stringList returns a string list of true positions in the bit vector.
// Mainly useful for debugging with smaller bit vectors.
func (b *bcVector) stringList() string {
	return b.bitVector().stringList()
}
