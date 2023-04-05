package bbhash_test

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"testing"
	"time"

	"github.com/relab/bbhash"
)

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
	salt := rand.New(rand.NewSource(99)).Uint64()

	tests := []struct {
		name   string
		fn     func(gamma float64, salt uint64, keys []uint64) (*bbhash.BBHash, error)
		keyMap map[uint64]uint64
	}{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel__", fn: bbhash.NewParallel},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bb, err := test.fn(2.0, salt, keys)
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
		})
	}
}

func TestManyKeys(t *testing.T) {
	sizes := []int{
		1000,
		10_000,
		100_000,
	}
	tests := []struct {
		name  string
		gamma float64
		seed  int
		fn    func(gamma float64, salt uint64, keys []uint64) (*bbhash.BBHash, error)
	}{
		{name: "Sequential", gamma: 1.0, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 1.0, seed: 123, fn: bbhash.NewParallel},
		{name: "Sequential", gamma: 1.5, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 1.5, seed: 123, fn: bbhash.NewParallel},
		{name: "Sequential", gamma: 2.0, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 2.0, seed: 123, fn: bbhash.NewParallel},
		{name: "Sequential", gamma: 2.5, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 2.5, seed: 123, fn: bbhash.NewParallel},
		{name: "Sequential", gamma: 3.0, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 3.0, seed: 123, fn: bbhash.NewParallel},
		{name: "Sequential", gamma: 5.0, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 5.0, seed: 123, fn: bbhash.NewParallel},
	}

	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		for _, size := range sizes {
			keys := generateKeys(size, tt.seed)
			t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, tt.gamma, size), func(t *testing.T) {
				bb, err := tt.fn(tt.gamma, salt, keys)
				if err != nil {
					t.Fatal(err)
				}
				keyMap := make(map[uint64]uint64)
				for keyIndex, key := range keys {
					hashIndex := bb.Find(key)
					checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
					if x, ok := keyMap[hashIndex]; ok {
						t.Fatalf("index %d already mapped to key %#x", hashIndex, x)
					}
					keyMap[hashIndex] = key
				}
			})
		}
	}
}

func TestSlow(t *testing.T) {
	// We only run this test if -timeout=0 is specified (ok == false).
	if _, ok := t.Deadline(); ok {
		// Find() is slow when checking more than 1 million keys
		t.Skip("Skipping test; use -timeout=0 to run it anyway")
	}
	// emit progress every 100k keys
	const progressInterval = 100_000
	sizes := []int{
		1_000_000,
		10_000_000,
		100_000_000,
	}
	tests := []struct {
		name  string
		gamma float64
		seed  int
		fn    func(gamma float64, salt uint64, keys []uint64) (*bbhash.BBHash, error)
	}{
		{name: "Sequential", gamma: 2.0, seed: 99, fn: bbhash.NewSequential},
		// {name: "Parallel__", gamma: 2.0, seed: 99, fn: bbhash.NewParallel},
	}

	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		for _, size := range sizes {
			keys := generateKeys(size, tt.seed)
			t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, tt.gamma, size), func(t *testing.T) {
				bb, err := tt.fn(tt.gamma, salt, keys)
				if err != nil {
					t.Fatal(err)
				}
				t.Log(bb)
				keyMap := make(map[uint64]uint64, size)
				start := time.Now()
				for keyIndex, key := range keys {
					if keyIndex%progressInterval == 0 {
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
						t.Fatalf("index %d already mapped to key %#x", hashIndex, x)
					}
					keyMap[hashIndex] = key
				}
			})
		}
	}
}

var bbSink *bbhash.BBHash

// BenchmarkNewBBHash benchmarks the creation of a new BBHash with sequential and parallel.
// Run with (it takes almost 30 minutes):
//
// go test -run x -bench BenchmarkNewBBHash -benchmem -cpuprofile cpu.prof -timeout=30m -count 10 > new.txt
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
	tests := []struct {
		name string
		fn   func(gamma float64, salt uint64, keys []uint64) (*bbhash.BBHash, error)
	}{
		{name: "Sequential", fn: bbhash.NewSequential},
		{name: "Parallel", fn: bbhash.NewParallel},
	}
	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		for _, size := range sizes {
			keys := generateKeys(size, 99)
			b.Run(fmt.Sprintf("name=%s/keys=%d", tt.name, size), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bbSink, _ = tt.fn(2.0, salt, keys)
				}
			})
		}
	}
}

func BenchmarkFind(b *testing.B) {
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
		bb, err := bbhash.NewSequential(tt.gamma, salt, keys)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(b *testing.B) {
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
