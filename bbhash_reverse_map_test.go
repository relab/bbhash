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
		{gamma: 2.0, size: 30},
		// {gamma: 2.0, size: 10000},
		// {gamma: 2.0, size: 100000},
		// {gamma: 2.0, size: 1000000},
		// Find() is too slow to check 10 million keys
	}

	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		keys := generateKeys(tt.size, 99)
		t.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(t *testing.T) {
			bb, reverseIndex, err := bbhash.NewWithReverseIndexNaive(tt.gamma, salt, keys)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(bb)
			for keyIndex, key := range keys {
				hashIndex := bb.Find(key)
				checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
				if reverseIndex[hashIndex] != key {
					t.Fatalf("index[%#x] = %#x, expected %#x", hashIndex, reverseIndex[hashIndex], key)
				}
			}
		})
		t.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(t *testing.T) {
			bb, err := bbhash.NewWithReverseIndex(tt.gamma, salt, keys)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(bb)
			errCnt, matchCnt := 0, 0
			for keyIndex, key := range keys {
				hashIndex := bb.Find(key)
				checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
				lookupKey, err := bb.Lookup(hashIndex)
				if err != nil {
					t.Fatal(err)
				}
				if lookupKey != key {
					errCnt++
					t.Errorf("Lookup(%d) = %#x, expected %#x", hashIndex, lookupKey, key)
				} else {
					matchCnt++
				}
			}
			t.Logf("Lookup() matched %d keys, %d keys did not match", matchCnt, errCnt)
			if errCnt > 0 {
				t.Fail()
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
		b.Run(fmt.Sprintf("Naive/gamma=%.1f/keys=%d", tt.gamma, tt.size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bbSink, revIndexSink, _ = bbhash.NewWithReverseIndexNaive(tt.gamma, salt, keys)
			}
		})
		b.Run(fmt.Sprintf(" Fast/gamma=%.1f/keys=%d", tt.gamma, tt.size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bbSink, _ = bbhash.NewWithReverseIndex(tt.gamma, salt, keys)
			}
		})
	}
}
