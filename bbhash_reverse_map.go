package bbhash

import (
	"errors"
	"fmt"
)

func NewWithReverseIndex(gamma float64, salt uint64, keys []uint64) (*BBHash, error) {
	if gamma <= 1.0 {
		gamma = 2.0
	}
	sz := uint64(len(keys))
	words := words(sz, gamma)
	bb := &BBHash{
		bits:       make([]*bitVector, 0, initialLevels),
		saltHash:   saltHash(salt),
		gamma:      gamma,
		revIndex:   make([]uint64, sz+1), // +1 for not-found
		current:    newBitVector(words),
		collisions: newBitVector(words),
		redo:       make([]uint64, 0, sz/2), // heuristic: only 1/2 of the keys will collide
	}
	if err := bb.computeWithReverseIndex(keys); err != nil {
		return nil, err
	}
	// clear intermediate results
	bb.current = nil
	bb.collisions = nil
	bb.redo = nil
	return bb, nil
}

func NewWithReverseIndexNaive(gamma float64, salt uint64, keys []uint64) (*BBHash, []uint64, error) {
	bb, err := NewSerial(gamma, salt, keys)
	if err != nil {
		return nil, nil, err
	}
	// bb.Length() should be len(keys) + 1 since 0 is reserved for not-found
	// Create reverse mapping
	revMap := make([]uint64, len(keys)+1)
	for _, k := range keys {
		revMap[bb.Find(k)] = k
	}
	return bb, revMap, nil
}

// Lookup returns the key for the given hash index. The hash index is expected to be in the range [1, len(keys)].
// If the hash index is 0, it means that the key was not in the original key set, and an error is returned.
// If the hash index is out of range or the reverse index was not initialized, an error is returned.
func (bb *BBHash) Lookup(hashIndex uint64) (uint64, error) {
	if bb.revIndex == nil {
		return 0, errors.New("reverse index not initialized")
	}
	if hashIndex == 0 {
		return 0, errors.New("key not found for hash index 0")
	}
	if hashIndex >= uint64(len(bb.revIndex)) {
		return 0, fmt.Errorf("hash index %d out of range", hashIndex)
	}
	return bb.revIndex[hashIndex], nil
}

// computeWithReverseIndex computes the minimal perfect hash for the given keys and creates the reverse index.
func (bb *BBHash) computeWithReverseIndex(keys []uint64) error {
	// Initializing the lvlRank to 1, since the 0 index is reserved for not-found.
	var lvlRank uint64 = 1
	bb.ranks = make([]uint64, initialLevels)

	allKeys := keys

	// loop exits when keys == nil, i.e., when there are no more keys to re-hash
	for lvl := uint(0); keys != nil; lvl++ {
		sz := bb.current.Size()
		// precompute the level hash to speed up the key hashing
		lvlHash := levelHash(bb.saltHash, uint64(lvl))

		// find colliding keys
		for _, k := range keys {
			i := keyHash(lvlHash, k) % sz
			if bb.current.IsSet(i) {
				// found one or more collisions at index i
				bb.collisions.Set(i)
				continue
			}
			bb.current.Set(i)
		}

		// assign non-colliding keys to the current level's bit vector
		for _, k := range keys {
			i := keyHash(lvlHash, k) % sz
			if bb.collisions.IsSet(i) {
				bb.redo = append(bb.redo, k)
				// unset the bit since there was a collision
				bb.current.Unset(i)
				continue
			}
			bb.current.Set(i)
		}

		// for _, k := range keys {
		// 	i := keyHash(lvlHash, k) % sz
		// 	if bb.current.IsSet(i) {
		// 		// we found the correct index position for this key
		// 		index := lvlRank + bb.current.Rank(i)
		// 		fmt.Printf("Comp(key=%#016x, pos=%2d) = (Level %d, L rank: %2d, P rank: %2d, Index: %2d)\n", k, i, lvl, lvlRank, bb.current.Rank(i), index)
		// 		if bb.revIndex[index] != 0 {
		// 			// index already used; key belongs to a higher level
		// 			fmt.Printf(" Dup(key=%#016x, pos=%2d) = (Level %d, L rank: %2d, P rank: %2d, Index: %2d)\n", k, i, lvl, lvlRank, bb.current.Rank(i), index)
		// 			continue
		// 		}
		// 		bb.revIndex[index] = k
		// 	}
		// }

		bb.ranks[lvl] = lvlRank
		lvlRank += bb.current.OnesCount()

		// move to next level and compute the set of keys to re-hash (that had collisions)
		keys = bb.nextLevel()

		if lvl > maxLevel {
			return fmt.Errorf("can't find minimal perfect hash after %d tries", lvl)
		}
	}
	// bb.computeLevelRanks()

	for _, k := range allKeys {
		for lvl, bv := range bb.bits {
			i := hash(bb.saltHash, uint64(lvl), k) % bv.Size()
			if bv.IsSet(i) {
				index := bb.ranks[lvl] + bv.Rank(i)
				bb.revIndex[index] = k
				fmt.Printf("Sind(key=%#016x) = (Index: %2d)\n", k, index)
				break
			}
		}
	}

	return nil
}
