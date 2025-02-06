package main

import (
	"fmt"
	"math/rand"

	"github.com/relab/bbhash"
)

// This program generates a lot of BBHashes and prints them, showing relevant statistics.
// In particular, it shows the number of bits per key, the number of bits per level, and
// an estimate of the false positive rate (it may be inaccurate). It is meant to give an
// indication of how the gamma parameter affects the size of the BBHash, and the false
// positive rate.
func main() {
	sizes := []int{
		1000,
		10_000,
		100_000,
	}
	for _, gamma := range []float64{1.1, 1.5, 1.7, 2.0, 2.5, 3.0, 5.0} {
		for _, size := range sizes {
			keys := generateKeys(size, 123)
			bb, err := bbhash.New(keys, bbhash.Gamma(gamma))
			if err != nil {
				panic(err)
			}
			fmt.Println(bb)
		}
	}
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
