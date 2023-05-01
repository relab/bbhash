// Package bbhash implements the BBHash algorithm for minimal perfect hash functions.
package bbhash

import (
	"fmt"
	"runtime"
	"sync"
)

// NewParallel2 creates a new BBHash for the given keys. The keys must be unique.
// This creates the BBHash using multiple goroutines.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
// The salt parameter is used to salt the hash function. Depending on your use case,
// you may use a cryptographic- or a pseudo-random number for the salt.
func NewParallel2(gamma float64, salt uint64, keys []uint64) (*BBHash, error) {
	if gamma <= 1.0 {
		gamma = 2.0
	}
	bb := &BBHash{
		bits:     make([]*bitVector, 0, initialLevels),
		saltHash: saltHash(salt),
	}
	if err := bb.computeParallel2(keys, gamma); err != nil {
		return nil, err
	}
	return bb, nil
}

// computeParallel2 computes the minimal perfect hash for the given keys.
func (bb *BBHash) computeParallel2(keys []uint64, gamma float64) error {
	sz := uint64(len(keys))
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	lvlVector := newBCVector(words(sz, gamma))
	wds := lvlVector.Words()
	ncpu := runtime.NumCPU()
	perCPUVectors := make([]*bcVector, ncpu)
	for i := 0; i < ncpu; i++ {
		perCPUVectors[i] = newBCVector(wds)
	}

	// loop exits when keys == nil, i.e., when there are no more keys to re-hash
	for lvl := 0; keys != nil; lvl++ {
		wds := lvlVector.Words()
		parts := split(keys, ncpu)

		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(bb.saltHash, uint64(lvl))

		var wg sync.WaitGroup
		wg.Add(len(parts))
		for j := 0; j < len(parts); j++ {
			if len(parts[j]) == 0 {
				wg.Done()
				continue
			}
			current := perCPUVectors[j]
			go func(j int) {
				current.reset(wds)
				// find colliding keys and possible bit vector positions for non-colliding keys
				for _, k := range parts[j] {
					h := keyHash(lvlHash, k)
					// update the bit and collision vectors for the current level
					current.Update(h)
				}
				// merge the current bit and collision vectors into the global bit and collision vectors
				lvlVector.Merge(current)
				wg.Done()
			}(j)
		}
		wg.Wait()

		// remove bit vector position assignments for colliding keys and add them to the redo set
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// unset the bit vector position for the current key if it collided
			if lvlVector.UnsetCollision(h) {
				redo = append(redo, k)
			}
		}

		// save the bit vector for the current level
		bb.bits = append(bb.bits, lvlVector.bitVector())

		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = redo
		sz = uint64(len(keys))
		if sz == 0 {
			break
		}
		redo = redo[:0]
		lvlVector.nextLevel(words(sz, gamma))

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()
	return nil
}
