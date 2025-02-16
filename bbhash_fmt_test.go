package bbhash

import (
	"math/rand"
	"testing"

	"github.com/relab/bbhash/internal/test"
)

// Default test parameters.
var (
	keySizes        = []int{1000, 10_000, 100_000, 1_000_000}
	partitionValues = []int{1, 4, 8, 16, 24, 32, 48, 64, 128}
	gammaValues     = []float64{1.0, 1.5, 2.0}
)

func TestString(t *testing.T) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				t.Run(test.Name("", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(t *testing.T) {
					bb2, _ := New(keys, Gamma(gamma), Partitions(partitions))
					if partitions == 1 {
						bb := bb2.SinglePartition()
						if size >= 1_000_000 {
							// Currently, it is too slow to calculate the false positive rate for 1M keys
							// See issue #21
							return
						}
						t.Logf("BBHash: %v", bb)
					}
					t.Logf("BBHash2: %v", bb2)
				})
			}
		}
	}
}

func TestBitsAndLengths(t *testing.T) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				t.Run(test.Name("", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(t *testing.T) {
					bb2, _ := New(keys, Gamma(gamma), Partitions(partitions))
					if partitions == 1 {
						// test also the BBHash implementation when there is a single partition
						testBitsAndLengths(t, size, bb2.SinglePartition())
					}
					testBitsAndLengths(t, size, bb2)
				})
			}
		}
	}
}

func testBitsAndLengths(t *testing.T, keyLen int, bb bbhashInternal) {
	t.Helper()

	wantMarshaledLength := bb.marshaledLength()

	// check that the actual marshaled data length is as expected;
	// if it is higher than expected, unnecessary allocations are being made
	data, _ := bb.MarshalBinary()
	gotMarshaledLength := len(data)
	if wantMarshaledLength != gotMarshaledLength {
		t.Errorf("marshaledLength() = %d, want %d", wantMarshaledLength, gotMarshaledLength)
	}

	wantWireBits := uint64(wantMarshaledLength * 8)
	gotWireBits := bb.wireBits()
	if wantWireBits != gotWireBits {
		t.Errorf("wireBits() = %d, want %d", gotWireBits, wantWireBits)
	}

	wantBPK := float64(wantWireBits) / float64(keyLen)
	gotBPK := bb.BitsPerKey()
	if wantBPK != gotBPK {
		t.Errorf("BitsPerKey() = %f, want %f", gotBPK, wantBPK)
	}

	wantEntries := uint64(keyLen)
	gotEntries := bb.entries()
	if wantEntries != gotEntries {
		t.Errorf("entries() = %d, want %d", gotEntries, wantEntries)
	}
}

type bbhashInternal interface {
	marshaledLength() int
	MarshalBinary() ([]byte, error)
	wireBits() uint64
	BitsPerKey() float64
	entries() uint64
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
