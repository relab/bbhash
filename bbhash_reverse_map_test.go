package bbhash_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/relab/bbhash"
)

func TestNewWithReverseMap(t *testing.T) {
	tests := []struct {
		gamma float64
		size  int
	}{
		{gamma: 2.0, size: 1000},
		{gamma: 2.0, size: 10000},
		{gamma: 2.0, size: 100000},
		// Construction and Find() is too slow to check 1 million keys (takes 30-40s).
	}
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		keys := generateKeys(tt.size, 99)
		t.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(t *testing.T) {
			for _, f := range []func() (*bbhash.BBHash, error){
				func() (*bbhash.BBHash, error) { return bbhash.NewWithReverseIndex(tt.gamma, salt, keys) },
			} {
				bb, err := f()
				if err != nil {
					t.Fatal(err)
				}
				t.Log(bb)
				for keyIndex, key := range keys {
					hashIndex := bb.Find(key)
					checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
					lookupKey, err := bb.Lookup(hashIndex)
					if err != nil {
						t.Fatal(err)
					}
					if lookupKey != key {
						t.Errorf("Lookup(%d) = %#x, expected %#x", hashIndex, lookupKey, key)
					}
				}
			}
		})
	}
}

var revIndexSink []uint64

func BenchmarkNewWithReverseMap(b *testing.B) {
	tests := []struct {
		gamma float64
		size  int
	}{
		{gamma: 2.0, size: 1000},
		{gamma: 2.0, size: 10000},
		{gamma: 2.0, size: 100000},
		{gamma: 2.0, size: 1000000},
	}
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		keys := generateKeys(tt.size, 99)
		b.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bbSink, _ = bbhash.NewWithReverseIndex(tt.gamma, salt, keys)
			}
		})
	}
}
