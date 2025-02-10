package bbhash

import (
	"fmt"
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

func TestBitVectorMarshalUnmarshalBinary(t *testing.T) {
	for _, words := range []uint64{1, 10, 100, 1000, 10000, 100000} {
		t.Run(fmt.Sprintf("words=%d", words), func(t *testing.T) {
			mbv := fillBitVector(words)
			data, err := mbv.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary() failed: %v", err)
			}

			var ubv bitVector
			if err := ubv.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary() failed: %v", err)
			}

			if !mbv.equal(ubv) {
				t.Errorf("bit vectors do not match")
			}
		})
	}
}

func TestBitVectorAppendBinary(t *testing.T) {
	for _, words := range []uint64{1, 10, 100, 1000, 10000, 100000} {
		t.Run(fmt.Sprintf("words=%d", words), func(t *testing.T) {
			abv := fillBitVector(words)
			data, err := abv.AppendBinary(nil)
			if err != nil {
				t.Fatalf("AppendBinary() failed: %v", err)
			}

			var ubv bitVector
			if err := ubv.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary() failed: %v", err)
			}

			if !abv.equal(ubv) {
				t.Errorf("bit vectors do not match")
			}
		})
	}
}

func BenchmarkBitVectorMarshalBinary(b *testing.B) {
	for _, words := range []uint64{1, 10, 100, 1000, 10000, 100000} {
		bv := fillBitVector(words)
		b.Run(fmt.Sprintf("words=%d", words), func(b *testing.B) {
			for b.Loop() {
				_, _ = bv.MarshalBinary()
			}
		})
	}
}

func BenchmarkBitVectorUnmarshalBinary(b *testing.B) {
	for _, words := range []uint64{1, 10, 100, 1000, 10000, 100000} {
		bv := fillBitVector(words)
		data, err := bv.MarshalBinary()
		if err != nil {
			b.Fatalf("MarshalBinary() failed: %v", err)
		}
		b.Run(fmt.Sprintf("words=%d", words), func(b *testing.B) {
			var bv bitVector
			for b.Loop() {
				_ = bv.UnmarshalBinary(data)
			}
		})
	}
}

func fillBitVector(words uint64) bitVector {
	bv := make(bitVector, words)
	for i := uint64(0); i < bv.size(); i++ {
		if i%2 == 0 {
			bv.set(i)
		}
	}
	return bv
}
