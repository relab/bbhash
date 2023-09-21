package bbhash

import (
	"fmt"
	"runtime"
	"sync"
)

// NewParallel creates a new BBHash for the given keys. The keys must be unique.
// This creates the BBHash using multiple goroutines.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func NewParallel(gamma float64, keys []uint64) (*BBHash, error) {
	if gamma <= 1.0 {
		gamma = 2.0
	}
	bb := newBBHash()
	if err := bb.computeParallel(gamma, keys); err != nil {
		return nil, err
	}
	return bb, nil
}

// computeParallel computes the minimal perfect hash for the given keys in parallel by sharding the keys.
func (bb *BBHash) computeParallel(gamma float64, keys []uint64) error {
	sz := len(keys)
	wds := words(sz, gamma)
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(wds)
	ncpu := runtime.NumCPU()
	var perCPUVectors []*bcVector
	if sz >= 40000 {
		perCPUVectors = make([]*bcVector, ncpu)
		for i := 0; i < ncpu; i++ {
			perCPUVectors[i] = newBCVector(wds)
		}
	}

	// loop exits when keys == nil, i.e., when there are no more keys to re-hash
	for lvl := uint(0); keys != nil; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(uint64(lvl))

		if sz < 40000 {
			for i := 0; i < len(keys); i++ {
				h := keyHash(lvlHash, keys[i])
				lvlVector.update(h)
			}
		} else {
			z := sz / ncpu
			r := sz % ncpu
			var wg sync.WaitGroup
			wg.Add(ncpu)
			for j := 0; j < ncpu; j++ {
				x := z * j
				y := x + z
				if j == ncpu-1 {
					y += r
				}
				if x == y {
					wg.Done()
					continue // no need to spawn a goroutine since there are no keys to process
				}
				current := perCPUVectors[j]
				go func() {
					current.reset(wds)
					// find colliding keys
					for _, k := range keys[x:y] {
						h := keyHash(lvlHash, k)
						// update the bit and collision vectors for the current level
						current.update(h)
					}
					wg.Done()
				}()
			}
			wg.Wait()
			// merge the per CPU bit and collision vectors into the global bit and collision vectors
			for _, v := range perCPUVectors {
				lvlVector.merge(v)
			}
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// unset the bit vector position for the current key if it collided
			if lvlVector.unsetCollision(h) {
				redo = append(redo, k)
			}
		}

		// save the current bit vector for the current level
		bb.bits = append(bb.bits, lvlVector.bitVector())

		sz = len(redo)
		if sz == 0 {
			break
		}
		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = redo
		redo = redo[:0]
		wds = words(sz, gamma)
		lvlVector.nextLevel(wds)

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()
	return nil
}
