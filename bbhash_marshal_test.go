package bbhash_test

import (
	"testing"

	"github.com/relab/bbhash"
)

func TestMarshalUnmarshalBBHash(t *testing.T) {
	size := 100000
	keys := generateKeys(size, 99)

	bb, err := bbhash.NewSequential(2.0, keys)
	if err != nil {
		t.Fatalf("Failed to create BBHash: %v", err)
	}

	// Create a map to hold the original Find() results
	originalHashIndexes := make(map[uint64]uint64, len(keys))
	for _, key := range keys {
		originalHashIndexes[key] = bb.Find(key)
	}

	data, err := bb.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal BBHash: %v", err)
	}

	newBB := &bbhash.BBHash{}
	if err = newBB.UnmarshalBinary(data); err != nil {
		t.Fatalf("Failed to unmarshal BBHash: %v", err)
	}

	// Validate that the unmarshalled BBHash returns the same Find() results
	for _, key := range keys {
		hashIndex := newBB.Find(key)
		if hashIndex != originalHashIndexes[key] {
			t.Fatalf("newBB.Find(%d) = %d, want %d", key, hashIndex, originalHashIndexes[key])
		}
	}
}
