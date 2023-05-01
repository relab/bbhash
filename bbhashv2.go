// Package bbhash implements the BBHash algorithm for minimal perfect hash functions.
package bbhash

import (
	"fmt"
)

// NewSequential2 creates a new BBHash for the given keys. The keys must be unique.
// This creates the BBHash in a single goroutine.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
// The salt parameter is used to salt the hash function. Depending on your use case,
// you may use a cryptographic- or a pseudo-random number for the salt.
func NewSequential2(gamma float64, salt uint64, keys []uint64) (*BBHash, error) {
	if gamma <= 1.0 {
		gamma = 2.0
	}
	bb := &BBHash{
		bits:     make([]*bitVector, 0, initialLevels),
		saltHash: saltHash(salt),
	}
	if err := bb.compute2(keys, gamma); err != nil {
		return nil, err
	}
	return bb, nil
}

// compute2 computes the minimal perfect hash for the given keys.
func (bb *BBHash) compute2(keys []uint64, gamma float64) error {
	sz := uint64(len(keys))
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(words(sz, gamma))

	// loop exits when keys == nil, i.e., when there are no more keys to re-hash
	for lvl := 0; keys != nil; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(bb.saltHash, uint64(lvl))

		// find colliding keys and possible bit vector positions for non-colliding keys
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// update the bit and collision vectors for the current level
			lvlVector.Update(h)
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// unset the bit vector position for the current key if it collided
			if lvlVector.UnsetCollision(h) {
				// keys to re-hash at next level : F in the paper
				redo = append(redo, k)
			}
		}

		// save the bit vector for the current level
		bb.bits = append(bb.bits, lvlVector.bitVector())

		sz = uint64(len(redo))
		if sz == 0 {
			break
		}
		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = redo
		redo = redo[:0]
		lvlVector.nextLevel(words(sz, gamma))

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()
	return nil
}