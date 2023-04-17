// Package bbhash implements the BBHash algorithm for minimal perfect hash functions.
package bbhash

import (
	"fmt"
	"runtime"
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
		gamma:    gamma,
	}
	if err := bb.compute2(keys); err != nil {
		return nil, err
	}
	return bb, nil
}

// compute2 computes the minimal perfect hash for the given keys.
func (bb *BBHash) compute2(keys []uint64) error {
	sz := uint64(len(keys))
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	ncpu := runtime.NumCPU()
	wds := words(sz, bb.gamma)
	fmt.Printf("ncpu=%d, len(keys)=%d, words=%d\n", ncpu, len(keys), wds)

	// loop exits when keys == nil, i.e., when there are no more keys to re-hash
	for lvl := 0; keys != nil; lvl++ {
		wds = words(uint64(len(keys)), bb.gamma)
		levelVector := newBCVector(wds)
		parts := split(keys, ncpu)

		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(bb.saltHash, uint64(lvl))

		for j := 0; j < ncpu; j++ {
			current := newBCVector(wds)

			// find colliding keys and possible bit vector positions for non-colliding keys
			for _, k := range parts[j] {
				h := keyHash(lvlHash, k)
				// fmt.Printf("lvl=%2d, keyHash=%#016x\n", lvl, h)
				// update the bit and collision vectors for the current level
				current.Update(h)
			}
			// fmt.Printf(" current.Size()=%d, OnesCount()=%d\n", current.Size(), current.bitVector().OnesCount())
			levelVector.Merge(current)
			// fmt.Printf("localVec.Size()=%d, OnesCount()=%d\n", levelVector.Size(), levelVector.bitVector().OnesCount())
		}

		for j := 0; j < ncpu; j++ {
			// remove bit vector position assignments for colliding keys and add them to the redo set
			for _, k := range parts[j] {
				h := keyHash(lvlHash, k)
				// unset the bit vector position for the current key if it collided
				if levelVector.UnsetCollision(h) {
					// fmt.Printf("bang: lvl=%2d, keyHash=%#016x, k=%#016x\n", lvl, h, k)
					redo = append(redo, k)
				}
			}
		}
		// save the current bit vector for the current cpu+level
		b := levelVector.bitVector()
		bb.bits = append(bb.bits, b)
		fmt.Printf("done: rank[%d]=%d, len(redo)=%d\n", len(bb.bits)-1, b.OnesCount(), len(redo))
		// fmt.Printf("vector: %s\n", b.String())

		if len(redo) == 0 {
			break
		}
		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = redo
		redo = redo[:0]

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()
	return nil
}
