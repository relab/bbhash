// Package bbhash implements the BBHash algorithm for minimal perfect hash functions.
package bbhash

import (
	"fmt"
)

const (
	// minimalGamma is the smallest allowed expansion factor for the bit vector.
	minimalGamma = 1.0
	// Heuristic: 32 levels should be enough for even very large key sets
	initialLevels = 32

	// Maximum number of attempts (level) at making a perfect hash function.
	// Per the paper, each successive level exponentially reduces the
	// probability of collision.
	maxLevel = 200
)

// BBHash represents a minimal perfect hash for a set of keys.
type BBHash struct {
	bits  []*bitVector
	ranks []uint64
}

func newBBHash() *BBHash {
	return &BBHash{
		bits: make([]*bitVector, 0, initialLevels),
	}
}

// NewSequential creates a new BBHash for the given keys. The keys must be unique.
// This creates the BBHash in a single goroutine.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func NewSequential(gamma float64, keys []uint64) (*BBHash, error) {
	gamma = max(gamma, minimalGamma)
	bb := newBBHash()
	if err := bb.compute(gamma, keys); err != nil {
		return nil, err
	}
	return bb, nil
}

// NewSequentialWithKeymap is similar to NewSequential, but in addition returns the reverse map.
func NewSequentialWithKeymap(gamma float64, keys []uint64) (*BBHash, []uint64, error) {
	gamma = max(gamma, minimalGamma)
	bb := newBBHash()
	keymap, err := bb.computeWithKeymap(gamma, keys)
	if err != nil {
		return nil, nil, err
	}
	return bb, keymap, nil
}

// Find returns a unique index representing the key in the minimal hash set.
//
// The return value is meaningful ONLY for keys in the original key set
// (provided at the time of construction of the minimal hash set).
//
// If the key is in the original key set, the return value is guaranteed to be
// in the range [1, len(keys)].
//
// If the key is not in the original key set, two things can happen:
// 1. The return value is 0, representing that the key was not in the original key set.
// 2. The return value is in the expected range [1, len(keys)], but is a false positive.
func (bb *BBHash) Find(key uint64) uint64 {
	for lvl, bv := range bb.bits {
		i := hash(uint64(lvl), key) % bv.size()
		if bv.isSet(i) {
			return bb.ranks[lvl] + bv.rank(i)
		}
	}
	return 0
}

// compute computes the minimal perfect hash for the given keys.
func (bb *BBHash) compute(gamma float64, keys []uint64) error {
	sz := len(keys)
	if sz == 0 {
		return fmt.Errorf("bbhash: compute: no keys")
	}

	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(words(sz, gamma))

	// loop exits when there are no more keys to re-hash (see break statement below)
	for lvl := 0; true; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(uint64(lvl))

		// find colliding keys and possible bit vector positions for non-colliding keys
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// update the bit and collision vectors for the current level
			lvlVector.update(h)
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// unset the bit vector position for the current key if it collided
			if lvlVector.unsetCollision(h) {
				// keys to re-hash at next level : F in the paper
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
		lvlVector.nextLevel(words(sz, gamma))

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()
	return nil
}

// computeWithKeymap is similar to compute(), but in addition returns the reverse keymap.
func (bb *BBHash) computeWithKeymap(gamma float64, keys []uint64) ([]uint64, error) {
	sz := len(keys)
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(words(sz, gamma))
	reverseMap := make([]uint64, len(keys)+1)
	levelKeysMap := make([][]uint64, initialLevels)
	// loop exits when there are no more keys to re-hash (see break statement below)
	for lvl := 0; true; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(uint64(lvl))

		// find colliding keys and possible bit vector positions for non-colliding keys
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// update the bit and collision vectors for the current level
			lvlVector.update(h)
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		levelKeys := make([]uint64, lvlVector.size())
		for _, k := range keys {
			h := keyHash(lvlHash, k)
			// unset the bit vector position for the current key if it collided
			if lvlVector.unsetCollision(h) {
				// keys to re-hash at next level : F in the paper
				redo = append(redo, k)
			} else {
				// keys for the current level used to construct the reverse map.
				// keys are ordered by their bit vector position to avoid sorting later.
				levelKeys[h%lvlVector.size()] = k
			}
		}
		levelKeysMap[lvl] = levelKeys

		// save the current bit vector for the current level
		bb.bits = append(bb.bits, lvlVector.bitVector())

		sz = len(redo)
		if sz == 0 {
			break
		}
		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = redo
		redo = redo[:0]
		lvlVector.nextLevel(words(sz, gamma))

		if lvl > maxLevel {
			return nil, fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	bb.computeLevelRanks()

	// Compute the reverse map
	index := 1
	for _, levelKeys := range levelKeysMap {
		for _, key := range levelKeys {
			if key != 0 {
				reverseMap[index] = key
				index++
			}
		}
	}
	return reverseMap, nil
}

// computeLevelRanks computes the total rank of each level.
// The total rank is the rank for all levels up to and including the current level.
func (bb *BBHash) computeLevelRanks() {
	// Initializing the rank to 1, since the 0 index is reserved for not-found.
	var rank uint64 = 1
	bb.ranks = make([]uint64, len(bb.bits))
	for l, bv := range bb.bits {
		bb.ranks[l] = rank
		rank += bv.onesCount()
	}
}
