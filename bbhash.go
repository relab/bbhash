// Package bbhash implements the BBHash algorithm for minimal perfect hash functions.
package bbhash

import (
	"fmt"

	"github.com/relab/bbhash/internal/fast"
)

const (
	// defaultGamma is the default expansion factor for the bit vector.
	defaultGamma = 2.0

	// minimalGamma is the smallest allowed expansion factor for the bit vector.
	minimalGamma = 0.5

	// Heuristic: 32 levels should be enough for even very large key sets
	initialLevels = 32

	// Maximum number of attempts (level) at making a perfect hash function.
	// Per the paper, each successive level exponentially reduces the
	// probability of collision.
	maxLevel = 200
)

// BBHash represents a minimal perfect hash for a set of keys.
type BBHash struct {
	bits       []*bitVector // bit vectors for each level
	ranks      []uint64     // total rank for each level
	reverseMap []uint64     // index -> key (only filled if needed)
}

func newBBHash(initialLevels int) *BBHash {
	return &BBHash{
		bits: make([]*bitVector, 0, initialLevels),
	}
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
		i := fast.Hash(uint64(lvl), key) % bv.size()
		if bv.isSet(i) {
			return bb.ranks[lvl] + bv.rank(i)
		}
	}
	return 0
}

// Key returns the key for the given index.
// The index must be in the range [1, len(keys)], otherwise 0 is returned.
func (bb *BBHash) Key(index uint64) uint64 {
	if bb.reverseMap == nil || index == 0 || int(index) >= len(bb.reverseMap) {
		return 0
	}
	return bb.reverseMap[index]
}

// compute computes the minimal perfect hash for the given keys.
func (bb *BBHash) compute(keys []uint64, gamma float64) error {
	sz := len(keys)
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(words(sz, gamma))

	// loop exits when there are no more keys to re-hash (see break statement below)
	for lvl := 0; true; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := fast.LevelHash(uint64(lvl))

		// find colliding keys and possible bit vector positions for non-colliding keys
		for _, k := range keys {
			h := fast.KeyHash(lvlHash, k)
			// update the bit and collision vectors for the current level
			lvlVector.update(h)
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		for _, k := range keys {
			h := fast.KeyHash(lvlHash, k)
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
func (bb *BBHash) computeWithKeymap(keys []uint64, gamma float64) error {
	sz := len(keys)
	redo := make([]uint64, 0, sz/2) // heuristic: only 1/2 of the keys will collide
	// bit vectors for current level : A and C in the paper
	lvlVector := newBCVector(words(sz, gamma))
	bb.reverseMap = make([]uint64, len(keys)+1)
	levelKeysMap := make([][]uint64, 0, len(bb.bits)) // number of initial levels = len(bb.bits)

	// loop exits when there are no more keys to re-hash (see break statement below)
	for lvl := 0; true; lvl++ {
		// precompute the level hash to speed up the key hashing
		lvlHash := fast.LevelHash(uint64(lvl))

		// find colliding keys and possible bit vector positions for non-colliding keys
		for _, k := range keys {
			h := fast.KeyHash(lvlHash, k)
			// update the bit and collision vectors for the current level
			lvlVector.update(h)
		}

		// remove bit vector position assignments for colliding keys and add them to the redo set
		levelKeys := make([]uint64, lvlVector.size())
		for _, k := range keys {
			h := fast.KeyHash(lvlHash, k)
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
		levelKeysMap = append(levelKeysMap, levelKeys)

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

	// Compute the reverse map
	index := 1
	for _, levelKeys := range levelKeysMap {
		for _, key := range levelKeys {
			if key != 0 {
				bb.reverseMap[index] = key
				index++
			}
		}
	}
	return nil
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

// enforce interface compliance
var (
	_ bbhash     = (*BBHash)(nil)
	_ reverseMap = (*BBHash)(nil)
)
