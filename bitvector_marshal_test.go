package bbhash

import "testing"

func TestMarshalUnmarshalBitVector(t *testing.T) {
	bv := newBitVector(3)
	bv.v = []uint64{42, 84, 168, 12, 3, 234, 12, 34, 12, 45, 4, 54, 44, 4, 43, 42, 23, 232, 35, 232, 67, 6, 2323, 9129123, 3232, 5, 45345, 345, 234, 23, 1232}

	data, err := bv.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	newBv := &bitVector{}
	err = newBv.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(newBv.v) != len(bv.v) {
		t.Fatalf("Length mismatch: got %d, want %d", len(newBv.v), len(bv.v))
	}

	for i, val := range bv.v {
		if newBv.v[i] != val {
			t.Fatalf("Value mismatch at index %d: got %d, want %d", i, newBv.v[i], val)
		}
	}
}
