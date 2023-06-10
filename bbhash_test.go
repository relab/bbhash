package bbhash_test

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"testing"
	"time"

	"github.com/relab/bbhash"
)

// This is only meant for testing, and should not be used for benchmarking.
type mphf interface{ Find(uint64) uint64 }

type variant[T mphf] struct {
	name string
	fn   func(gamma float64, salt uint64, keys []uint64) (T, error)
	fn2  func(gamma float64, partitionSize int, salt uint64, keys []uint64) (T, error)
}

func runMPHFTest[T mphf](t *testing.T, tt variant[T], keys []uint64, gamma float64) {
	t.Helper()
	// emit progress every 100k keys
	const progressInterval = 100_000
	size := len(keys)
	logProgress := size > 2*progressInterval
	salt := rand.New(rand.NewSource(99)).Uint64()
	t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, gamma, size), func(t *testing.T) {
		var bb T
		var err error
		if tt.fn != nil {
			bb, err = tt.fn(gamma, salt, keys)
		} else if tt.fn2 != nil {
			bb, err = tt.fn2(gamma, 20, salt, keys)
		} else {
			t.Fatal("no function to test")
		}
		if err != nil {
			t.Fatal(err)
		}
		if logProgress {
			fmt.Println(bb)
		}
		keyMap := make(map[uint64]uint64)
		start := time.Now()
		for keyIndex, key := range keys {
			if logProgress && keyIndex%progressInterval == 0 {
				duration := time.Since(start)
				if duration > time.Second {
					duration = duration.Truncate(time.Second)
					expectedTimeToFinish := time.Duration(size/progressInterval) * duration
					t.Logf("Progress (keyIndex=%9d) Duration: %s Expect to finish in %s", keyIndex, duration, expectedTimeToFinish)
				}
				start = time.Now()
			}

			hashIndex := bb.Find(key)
			checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
			if x, ok := keyMap[hashIndex]; ok {
				t.Errorf("index %d already mapped to key %#x", hashIndex, x)
			}
			keyMap[hashIndex] = key
		}
	})
}

func TestSimple(t *testing.T) {
	someStarWarsCharacters := []string{
		"4-LOM",
		"Admiral Thrawn",
		"Senator Bail Organa",
		"Ben Skywalker",
		"Bib Fortuna",
		"Boba Fett",
		"C-3PO",
		"Cad Bane",
		"Cade Skywalker",
		"Captain Rex",
		"Chewbacca",
		"Clone Commander Cody",
		"Darth Vader",
		"General Grievous",
		"General Veers",
		"Greedo",
		"Han Solo",
		"IG 88",
		"Jabba The Hutt",
		"Luke Skywalker",
		"Mara Jade",
		"Mission Vao",
		"Obi-Wan Kenobi",
		"Princess Leia",
		"PROXY",
		"Qui-Gon Jinn",
		"R2-D2",
		"Revan",
		"Wedge Antilles",
		"Yoda",
	}
	keys := make([]uint64, len(someStarWarsCharacters))
	for i, s := range someStarWarsCharacters {
		keys[i] = fnvHash(s)
	}
	tests := []variant[*bbhash.BBHash]{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel__", fn: bbhash.NewParallel},
	}
	tests2 := []variant[*bbhash.BBHash2]{
		{name: "Parallel2_", fn2: bbhash.NewParallel2},
	}
	for _, tt := range tests {
		runMPHFTest(t, tt, keys, 2.0)
	}
	for _, tt := range tests2 {
		runMPHFTest(t, tt, keys, 2.0)
	}
}

func TestManyKeys(t *testing.T) {
	tests := []variant[*bbhash.BBHash]{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel__", fn: bbhash.NewParallel},
	}
	tests2 := []variant[*bbhash.BBHash2]{
		{name: "Parallel2_", fn2: bbhash.NewParallel2},
	}
	sizes := []int{
		1000,
		10_000,
		100_000,
	}
	const seed = 123
	for _, gamma := range []float64{1.1, 1.5, 2.0, 2.5, 3.0, 5.0} {
		for _, size := range sizes {
			keys := generateKeys(size, seed)
			for _, tt := range tests {
				runMPHFTest(t, tt, keys, gamma)
			}
			for _, tt := range tests2 {
				runMPHFTest(t, tt, keys, gamma)
			}
		}
	}
}

func TestSlow(t *testing.T) {
	// We only run this test if -timeout=0 is specified (ok == false).
	if _, ok := t.Deadline(); ok {
		// Find() is slow when checking more than 1 million keys
		t.Skip("Skipping test; use -timeout=0 to run it anyway")
	}
	sizes := []int{
		1_000_000,
		10_000_000,
		100_000_000,
	}
	tests := []variant[*bbhash.BBHash]{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel__", fn: bbhash.NewParallel},
	}
	tests2 := []variant[*bbhash.BBHash2]{
		{name: "Parallel2_", fn2: bbhash.NewParallel2},
	}
	for _, size := range sizes {
		keys := generateKeys(size, 99)
		for _, tt := range tests {
			runMPHFTest(t, tt, keys, 2.0)
		}
		for _, tt := range tests2 {
			runMPHFTest(t, tt, keys, 2.0)
		}
	}
}

// Run with:
//
//	go test -run TestFalsePositiveRate
//
// Avoid the -v argument to make it more readable.
func TestFalsePositiveRate(t *testing.T) {
	sizes := []int{
		1000,
		10_000,
		100_000,
	}
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, gamma := range []float64{1.1, 1.5, 1.7, 2.0, 2.5, 3.0, 5.0} {
		for _, size := range sizes {
			keys := generateKeys(size, 123)
			t.Run(fmt.Sprintf("gamma=%0.1f/keys=%d", gamma, size), func(t *testing.T) {
				bb, err := bbhash.NewSequential(gamma, salt, keys)
				if err != nil {
					t.Fatal(err)
				}
				keyMap := make(map[uint64]uint64)
				for keyIndex, key := range keys {
					hashIndex := bb.Find(key)
					checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
					if x, ok := keyMap[hashIndex]; ok {
						t.Errorf("index %d already mapped to key %#x", hashIndex, x)
					}
					keyMap[hashIndex] = key
				}
				fmt.Println(bb)
			})
		}
	}
}

var bbSink mphf

// BenchmarkNewBBHash benchmarks the creation of a new BBHash with sequential and parallel.
// Run with a large timeout (on my M2 Max it took 38 minutes):
//
// go test -run x -bench BenchmarkNewBBHash -benchmem -timeout=60m -count 20 > new.txt
//
// Then compare with:
//
// benchstat -col /name new.txt
func BenchmarkNewBBHash(b *testing.B) {
	sizes := []int{
		1000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
		1_000_000_000,
	}
	gammaValues := []float64{
		1.1,
		1.3,
		1.5,
		2.0,
	}
	tests := []variant[*bbhash.BBHash]{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel", fn: bbhash.NewParallel},
	}
	tests2 := []variant[*bbhash.BBHash2]{
		{name: "Parallel3", fn2: bbhash.NewParallel2},
	}
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, size := range sizes {
		keys := generateKeys(size, 99)
		for _, tt := range tests {
			for _, gamma := range gammaValues {
				bb, _ := tt.fn(2.0, salt, keys)
				lvls := bb.Levels()
				b.Run(fmt.Sprintf("name=%s/gamma=%.1f/bpk=%.2f/Levels=%d/keys=%d", tt.name, gamma, bb.BitsPerKey(), lvls, size), func(b *testing.B) {
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						bbSink, _ = tt.fn(2.0, salt, keys)
					}
				})
			}
		}
		for _, tt := range tests2 {
			for _, gamma := range gammaValues {
				for partSize := 2; partSize <= 512; partSize += 2 {
					bb, _ := tt.fn2(gamma, partSize, salt, keys)
					max, min := bb.MaxMinLevels()
					b.Run(fmt.Sprintf("name=%s/gamma=%.1f/bpk=%.2f/Levels=%d,%d/PartitionSize=%d/keys=%d", tt.name, gamma, bb.BitsPerKey(), max, min, partSize, size), func(b *testing.B) {
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							bbSink, _ = tt.fn2(gamma, partSize, salt, keys)
						}
					})
				}
			}
		}
	}
}

func BenchmarkFind(b *testing.B) {
	sizes := []int{
		1000,
		10_000,
		100_000,
		1_000_000,
		// 10_000_000,
		// 100_000_000,
		// 1_000_000_000,
	}
	tests := []variant[*bbhash.BBHash]{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel", fn: bbhash.NewParallel},
	}
	tests2 := []variant[*bbhash.BBHash2]{
		{name: "Parallel2", fn2: bbhash.NewParallel2},
	}
	const gamma = 2.0
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, size := range sizes {
		keys := generateKeys(size, 99)
		for _, tt := range tests {
			bb, err := tt.fn(gamma, salt, keys)
			if err != nil {
				b.Fatal(err)
			}
			b.Run(fmt.Sprintf("name=%s/gamma=%.1f/keys=%d", tt.name, gamma, size), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, k := range keys {
						if bb.Find(k) == 0 {
							b.Fatalf("can't find the key: %#x", k)
						}
					}
				}
			})
		}
		for _, tt := range tests2 {
			bb, err := tt.fn(gamma, salt, keys)
			if err != nil {
				b.Fatal(err)
			}
			b.Run(fmt.Sprintf("name=%s/gamma=%.1f/keys=%d", tt.name, gamma, size), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, k := range keys {
						if bb.Find(k) == 0 {
							b.Fatalf("can't find the key: %#x", k)
						}
					}
				}
			})
		}
	}
}

func checkKey(t *testing.T, keyIndex int, key, entries, hashIndex uint64) {
	t.Helper()
	if hashIndex == 0 {
		t.Fatalf("can't find key: %#x", key)
	}
	if hashIndex > entries {
		t.Fatalf("key %d <%#x> mapping %d out-of-bounds", keyIndex, key, hashIndex)
	}
}

// fnvHash hashes a string to a uint64.
func fnvHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
