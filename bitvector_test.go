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

func TestBitVectorMarshalUnmarshalSmall(t *testing.T) {
	tests := []struct {
		name    string
		bv      bitVector
		wantErr bool
	}{
		{name: "EmptyVector", bv: bitVector{}, wantErr: true},
		{name: "SingleWord", bv: bitVector{0x1234567890abcdef}},
		{name: "TwoWords", bv: bitVector{0xdeadbeefcafebabe, 0xfeedface11223344}},
		{name: "TenWords", bv: bitVector(make([]uint64, 10))},
		{name: "ManyWords", bv: bitVector{42, 84, 168, 12, 3, 234, 12, 34, 12, 45, 4, 54, 44, 4, 43, 42, 23, 232, 35, 232, 67, 6, 2323, 9129123, 3232, 5, 45345, 345, 234, 23, 1232}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.bv.MarshalBinary()
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			var decoded bitVector
			err = decoded.UnmarshalBinary(data)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}
			// Should not get an error
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			if len(decoded) != len(tc.bv) {
				t.Errorf("Length mismatch: got %d, want %d", len(decoded), len(tc.bv))
			}
			if !tc.bv.equal(decoded) {
				t.Errorf("Unmarshaled bitVector does not match original:\nOriginal: %v\n Decoded: %v", tc.bv, decoded)
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
