package bbhash_test

import (
	"fmt"
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

// Run with:
// go test -run x -bench BenchmarkMarshalBinaryBBHash -benchmem
func BenchmarkMarshalBinaryBBHash(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			b.Run(fmt.Sprintf("gamma=%.1f/keys=%d", gamma, size), func(b *testing.B) {
				bb, _ := bbhash.NewSequential(gamma, keys)
				bpk := bb.BitsPerKey()
				dataLen := 0
				b.ResetTimer()
				for range b.N {
					d, err := bb.MarshalBinary()
					if err != nil {
						b.Fatalf("Failed to marshal BBHash: %v", err)
					}
					dataLen += len(d)
				}
				// This metric is always the same for a given set of keys.
				b.ReportMetric(bpk, "bits/key")
				// This metric correspond to bits/key: dataLen*8/len(keys)
				b.ReportMetric(float64(dataLen)/float64(b.N), "bytes/msg")
			})
		}
	}
}

// Run with:
// go test -run x -bench BenchmarkUnmarshalBinaryBBHash -benchmem
func BenchmarkUnmarshalBinaryBBHash(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			b.Run(fmt.Sprintf("gamma=%.1f/keys=%d", gamma, size), func(b *testing.B) {
				bb, _ := bbhash.NewSequential(gamma, keys)
				bpk := bb.BitsPerKey()
				d, err := bb.MarshalBinary()
				if err != nil {
					b.Fatalf("Failed to marshal BBHash: %v", err)
				}
				dataLen := len(d)
				b.ResetTimer()
				for range b.N {
					newBB := &bbhash.BBHash{}
					if err = newBB.UnmarshalBinary(d); err != nil {
						b.Fatalf("Failed to unmarshal BBHash: %v", err)
					}
					newBpk := newBB.BitsPerKey()
					if newBpk != bpk {
						b.Fatalf("newBB.BitsPerKey() = %f, want %f", newBpk, bpk)
					}
				}
				// This metric is always the same for a given set of keys.
				b.ReportMetric(bpk, "bits/key")
				// This metric correspond to bits/key: dataLen*8/len(keys)
				b.ReportMetric(float64(dataLen), "bytes/msg")
			})
		}
	}
}
