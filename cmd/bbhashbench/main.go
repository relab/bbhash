package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/relab/bbhash"
)

func main() {
	var (
		name       = flag.String("name", "seq", "name of the mphf to benchmark (seq, par, par2)")
		gamma      = flag.Float64("gamma", 2.0, "gamma parameter")
		salt       = flag.Uint64("salt", 99, "salt parameter")
		keys       = flag.Int("keys", 1000, "number of keys to use")
		partitions = flag.Int("partitions", 1, "number of partitions to use (for par2 only)")
	)
	flag.Parse()
	switch *name {
	case "seq":
		runSequential(*keys, *gamma, *salt)
	case "par":
		runParallel(*keys, *gamma, *salt)
	case "par2":
		runParallel2(*keys, *partitions, *gamma, *salt)
	default:
		panic("unknown mphf name")
	}
}

func runSequential(numKeys int, gamma float64, salt uint64) {
	keys := generateKeys(numKeys, 99)
	start := time.Now()
	bb, err := bbhash.NewSequential(gamma, salt, keys)
	elapsed := time.Since(start)
	if err != nil {
		panic(err)
	}
	fmt.Printf("mphf=%s/gamma=%0.1f/keys=%d/salt=%d/elapsed=%s\n", "Sequential", gamma, numKeys, salt, elapsed)
	fmt.Println(bb)
}

func runParallel(numKeys int, gamma float64, salt uint64) {
	keys := generateKeys(numKeys, 99)
	start := time.Now()
	bb, err := bbhash.NewParallel(gamma, salt, keys)
	elapsed := time.Since(start)
	if err != nil {
		panic(err)
	}
	fmt.Printf("mphf=%s/gamma=%0.1f/keys=%d/salt=%d/elapsed=%s\n", "Parallel", gamma, numKeys, salt, elapsed)
	fmt.Println(bb)
}

func runParallel2(numKeys, numPartitions int, gamma float64, salt uint64) {
	keys := generateKeys(numKeys, 99)
	start := time.Now()
	bb, err := bbhash.NewParallel2(gamma, numPartitions, salt, keys)
	elapsed := time.Since(start)
	if err != nil {
		panic(err)
	}
	fmt.Printf("mphf=%s/gamma=%0.1f/keys=%d/partitions=%d/salt=%d/elapsed=%s\n", "Parallel2", gamma, numKeys, numPartitions, salt, elapsed)
	fmt.Println(bb)
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
