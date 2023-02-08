package bbhash

import (
	"testing"
)

func TestSetIsSet(t *testing.T) {
	const words = 50
	bv := newBitVector(words)

	for i := uint64(0); i < bv.Size(); i++ {
		if bv.IsSet(i) {
			t.Errorf("IsSet(%d) = true, expected false", i)
		}
	}
	for i := uint64(0); i < bv.Size(); i++ {
		bv.Set(i)
		if !bv.IsSet(i) {
			t.Errorf("IsSet(%d) = false, expected true", i)
		}
	}

	bv.Reset(words)

	for i := uint64(0); i < bv.Size(); i++ {
		if bv.IsSet(i) {
			t.Errorf("IsSet(%d) = true, expected false", i)
		}
	}
	for i := uint64(0); i < bv.Size(); i++ {
		if i%2 == 0 {
			continue
		}
		bv.Set(i)
	}

	for i := uint64(0); i < bv.Size(); i++ {
		isSet := bv.IsSet(i)
		if i%2 == 0 {
			if isSet {
				t.Errorf("IsSet(%d) = true, expected false", i)
			}
		} else {
			if !isSet {
				t.Errorf("IsSet(%d) = false, expected true", i)
			}
		}
	}
}
