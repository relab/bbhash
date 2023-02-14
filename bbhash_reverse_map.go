package bbhash

import (
	"errors"
	"fmt"
)

// NewWithReverseIndex creates a new BBHash with a reverse index.
// See NewSerial for details on the parameters.
func NewWithReverseIndex(gamma float64, salt uint64, keys []uint64) (*BBHash, error) {
	bb, err := NewSerial(gamma, salt, keys)
	if err != nil {
		return nil, err
	}
	// Create reverse index
	revMap := make([]uint64, len(keys)+1) // +1 since 0 is reserved for not-found
	for _, k := range keys {
		revMap[bb.Find(k)] = k
	}
	bb.revIndex = revMap
	return bb, nil
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
