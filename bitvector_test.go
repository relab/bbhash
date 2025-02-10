package bbhash

import (
	"testing"
)

func TestSetIsSet(t *testing.T) {
	const words = 50
	bv := make(bitVector, words)

	for i := uint64(0); i < bv.size(); i++ {
		if bv.isSet(i) {
			t.Errorf("IsSet(%d) = true, expected false", i)
		}
	}
	for i := uint64(0); i < bv.size(); i++ {
		bv.set(i)
		if !bv.isSet(i) {
			t.Errorf("IsSet(%d) = false, expected true", i)
		}
	}

	bv.reset(words)

	for i := uint64(0); i < bv.size(); i++ {
		if bv.isSet(i) {
			t.Errorf("IsSet(%d) = true, expected false", i)
		}
	}
	for i := uint64(0); i < bv.size(); i++ {
		if i%2 == 0 {
			continue
		}
		bv.set(i)
	}

	for i := uint64(0); i < bv.size(); i++ {
		isSet := bv.isSet(i)
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
