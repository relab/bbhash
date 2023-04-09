package bbhash

import (
	"sync"
)

// bcVector represents a combined bit and collision vector.
type bcVector struct {
	v  []uint64
	c  []uint64
	mu sync.Mutex
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

func (b *bcVector) reset(words uint64) {
	b.c = b.c[:words]
	b.v = b.v[:words]
	for i := range b.c {
		b.c[i] = 0
		b.v[i] = 0
	}
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

// Words returns the number of 64-bit words this bit vector has allocated.
func (b *bcVector) Words() uint64 {
	return uint64(len(b.v))
}

// Update sets the bit for the given hash h, and records a collision if the bit was already set.
// The bit position is determined by h modulo the size of the vector.
func (b *bcVector) Update(h uint64) {
	x := (h % b.Size()) / 64
	y := uint64(1 << (h & 63))
	if b.v[x]&y != 0 {
		// found one or more collisions at index i; update collision vector
		b.c[x] |= y
		return
	}
	// no collisions at index i; set bit
	b.v[x] |= y
}

// UnsetCollision returns true if hash h's bit has a collision.
// The vector is also unset for hash h's bit position.
// The bit position is determined by h modulo the size of the vector.
func (b *bcVector) UnsetCollision(h uint64) bool {
	x := (h % b.Size()) / 64
	y := uint64(1 << (h & 63))
	if b.c[x]&y != 0 {
		// found collision at index i; unset bit
		b.v[x] &^= y
		return true
	}
	// no collisions at index i
	return false
}

// Merge merges the local bcVector into the this bcVector.
func (b *bcVector) Merge(local *bcVector) {
	// Below v (b.v) refers to the existing global bit vector, and lv (local.v) refers
	// to the bit vector to be merged into the global bit vector.
	//
	//   v   lv   AND   OR   New v   Collision   Note
	//   0    0    0     0     0         0       Not set, no collision
	//   0    1    0     1     1         0       Not set, update v, no collision
	//   1    0    0     1     1         0       Already set, no collision
	//   1    1    1     1     1         1       Leave it set, collision
	//
	// If v&lv is 0, then there is no collision for the corresponding bit-pairs in the two bit vectors.
	// However, the lc vector may still have collisions, so we merge lc into the global collision vector
	// if lc is not 0.
	//
	// Note: only b.v and b.c needs to be locked.
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.v {
		v := b.v[i]
		lv := local.v[i]
		lc := local.c[i]
		c := v&lv | lc
		if c != 0 {
			b.c[i] |= c
		}
		b.v[i] |= lv
	}
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
