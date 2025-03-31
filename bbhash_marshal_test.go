package bbhash

import (
	"testing"

	"github.com/relab/bbhash/internal/test"
)

func TestMarshalUnmarshalSingleBBHash(t *testing.T) {
	size := 100000
	keys := generateKeys(size, 99)

	bb2, err := New(keys, Gamma(2.0))
	if err != nil {
		t.Fatalf("Failed to create BBHash: %v", err)
	}
	bb := bb2.SinglePartition()

	// Create a map to hold the original Find() results
	originalHashIndexes := make(map[uint64]uint64, len(keys))
	for _, key := range keys {
		originalHashIndexes[key] = bb.Find(key)
	}

	data, err := bb.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal BBHash: %v", err)
	}

	newBB := &SingleBBHash{}
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

func TestMarshalUnmarshalBBHashEmpty(t *testing.T) {
	bb := &SingleBBHash{}
	data, err := bb.MarshalBinary()
	if err == nil {
		t.Errorf("MarshalBinary() should have failed")
	}
	newBB := &SingleBBHash{}
	if err = newBB.UnmarshalBinary(data); err == nil {
		t.Errorf("UnmarshalBinary() should have failed")
	}
}

func TestMarshalUnmarshalBBHash(t *testing.T) {
	testCases := []struct {
		size       int
		partitions int
	}{
		{size: 100, partitions: 1},
		{size: 100, partitions: 2},
		{size: 1000, partitions: 2},
		{size: 1000, partitions: 4},
		{size: 10000, partitions: 4},
		{size: 10000, partitions: 8},
		{size: 100000, partitions: 8},
		{size: 100000, partitions: 16},
		{size: 1000000, partitions: 16},
		{size: 1000000, partitions: 32},
		{size: 1000000, partitions: 64},
	}

	for _, tc := range testCases {
		t.Run(test.Name("", []string{"keys", "partitions"}, tc.size, tc.partitions), func(t *testing.T) {
			keys := generateKeys(tc.size, 98)

			bb, err := New(keys, Partitions(tc.partitions))
			if err != nil {
				t.Fatalf("Failed to create BBHash: %v", err)
			}

			// Store original Find() results
			originalHashIndexes := make(map[uint64]uint64, len(keys))
			for _, key := range keys {
				originalHashIndexes[key] = bb.Find(key)
			}

			t.Logf("Original BBHash: %v", bb)

			data, err := bb.MarshalBinary()
			if err != nil {
				t.Fatalf("Failed to marshal BBHash: %v", err)
			}

			newBB := &BBHash{}
			if err = newBB.UnmarshalBinary(data); err != nil {
				t.Fatalf("Failed to unmarshal BBHash: %v", err)
			}

			// Validate that the unmarshalled BBHash produces the same Find() results
			for _, key := range keys {
				hashIndex := newBB.Find(key)
				if hashIndex != originalHashIndexes[key] {
					t.Fatalf("Mismatch: newBB.Find(%d) = %d, want %d", key, hashIndex, originalHashIndexes[key])
				}
			}
		})
	}
}

// Run with:
// go test -run x -bench BenchmarkBBHashMarshalBinary -benchmem
func BenchmarkSingleBBHashMarshalBinary(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			b.Run(test.Name("", []string{"gamma", "keys"}, gamma, size), func(b *testing.B) {
				bb2, _ := New(keys, Gamma(gamma))
				bb := bb2.SinglePartition()
				bpk := bb.BitsPerKey()

				data, err := bb.MarshalBinary()
				if err != nil {
					b.Fatalf("Failed to marshal BBHash: %v", err)
				}
				marshaledSize := len(data)

				b.ResetTimer()
				for b.Loop() {
					_, _ = bb.MarshalBinary()
				}
				// This metric is always the same for a given set of keys.
				b.ReportMetric(bpk, "bits/key")
				b.ReportMetric(float64(marshaledSize), "Bytes")
			})
		}
	}
}

// Run with:
// go test -run x -bench BenchmarkBBHashUnmarshalBinary -benchmem
func BenchmarkSingleBBHashUnmarshalBinary(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			b.Run(test.Name("", []string{"gamma", "keys"}, gamma, size), func(b *testing.B) {
				bb2, _ := New(keys, Gamma(gamma))
				bb := bb2.SinglePartition()
				bpk := bb.BitsPerKey()

				data, err := bb.MarshalBinary()
				if err != nil {
					b.Fatalf("Failed to marshal BBHash: %v", err)
				}
				marshaledSize := len(data)

				newBB := &BBHash{}
				if err = newBB.UnmarshalBinary(data); err != nil {
					b.Fatalf("Failed to unmarshal BBHash: %v", err)
				}
				newBpk := newBB.BitsPerKey()
				if newBpk != bpk {
					b.Fatalf("newBB.BitsPerKey() = %f, want %f", newBpk, bpk)
				}

				b.ResetTimer()
				for b.Loop() {
					_ = newBB.UnmarshalBinary(data)
				}
				// This metric is always the same for a given set of keys.
				b.ReportMetric(bpk, "bits/key")
				b.ReportMetric(float64(marshaledSize), "Bytes")
			})
		}
	}
}

// Run with:
// go test -run x -bench BenchmarkBBHashMarshalBinary -benchmem
func BenchmarkBBHashMarshalBinary(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(test.Name("", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(b *testing.B) {
					bb, _ := New(keys, Gamma(gamma), Partitions(partitions))
					bpk := bb.BitsPerKey()

					data, err := bb.MarshalBinary()
					if err != nil {
						b.Fatalf("Failed to marshal BBHash: %v", err)
					}
					marshaledSize := len(data)

					b.ResetTimer()
					for b.Loop() {
						_, _ = bb.MarshalBinary()
					}
					// This metric is always the same for a given set of keys.
					b.ReportMetric(bpk, "bits/key")
					b.ReportMetric(float64(marshaledSize), "Bytes")
				})
			}
		}
	}
}

// Run with:
// go test -run x -bench BenchmarkBBHashUnmarshalBinary -benchmem
func BenchmarkBBHashUnmarshalBinary(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(test.Name("", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(b *testing.B) {
					bb, _ := New(keys, Gamma(gamma), Partitions(partitions))
					bpk := bb.BitsPerKey()

					data, err := bb.MarshalBinary()
					if err != nil {
						b.Fatalf("Failed to marshal BBHash: %v", err)
					}
					marshaledSize := len(data)

					newBB := &BBHash{}
					if err = newBB.UnmarshalBinary(data); err != nil {
						b.Fatalf("Failed to unmarshal BBHash: %v", err)
					}
					newBpk := newBB.BitsPerKey()
					if newBpk != bpk {
						b.Fatalf("newBB.BitsPerKey() = %f, want %f", newBpk, bpk)
					}

					b.ResetTimer()
					for b.Loop() {
						_ = newBB.UnmarshalBinary(data)
					}
					// This metric is always the same for a given set of keys.
					b.ReportMetric(bpk, "bits/key")
					b.ReportMetric(float64(marshaledSize), "Bytes")
				})
			}
		}
	}
}

// This is a fast deterministic benchmark that only measures the message length (BBHash length)
// and number of bits per key for different key sizes, gamma values and number of partitions.
//
// Run with:
// go test -run x -bench BenchmarkBBHashBitsPerKey
func BenchmarkBBHashBitsPerKey(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(test.Name("", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(b *testing.B) {
					//lint:ignore SA3001 only need to run once since we are measuring deterministic output
					b.N = 1
					// Stop the benchmark timer since we measure only the bits/key calculation
					b.StopTimer()

					bb, _ := New(keys, Gamma(gamma), Partitions(partitions))
					bpk := bb.BitsPerKey()
					data, _ := bb.MarshalBinary()
					marshaledSize := len(data)

					// This metric is always the same for a given set of keys.
					b.ReportMetric(bpk, "bits/key")
					b.ReportMetric(float64(marshaledSize), "Bytes")

					// Suppress the default metrics
					b.ReportMetric(0, "ns/op")
					b.ReportMetric(0, "B/op")
					b.ReportMetric(0, "allocs/op")
				})
			}
		}
	}
}
