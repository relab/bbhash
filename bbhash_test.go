package bbhash_test

import (
	"flag"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/relab/bbhash"
)

// Default benchmark parameters.
var (
	keySizes        = []int{1000, 10_000, 100_000, 1_000_000}
	longKeySizes    = []int{10_000_000, 100_000_000, 1_000_000_000}
	partitionValues = []int{1, 4, 8, 16, 24, 32, 48, 64, 128}
	gammaValues     = []float64{1.0, 1.5, 2.0}
)

// TestMain parses command-line flags to set the key sizes, partition values, and gamma values.
//
// To run all main benchmarks:
//
//	go test -run x -bench Benchmark -benchmem -timeout=0 -gamma=1,1.5,2 -partitions=1,2,4,8 -keys=1000,10000
//	go test -run x -bench Benchmark -benchmem -timeout=0 -gamma=1,1.5,2 -partitions=1,2,4,8 -keys=long
//
// To run specific benchmarks:
//
//	go test -run x -bench BenchmarkBBHashNew -benchmem -timeout=0 -count 2 -gamma=1.5,2 -partitions=1,2,4,8 -keys=1000,10000
//	go test -run x -bench BenchmarkBBhashFind -benchmem -timeout=0 -count 2 -gamma=1.5,2 -partitions=1,2,4,8 -keys=1000,10000
//	go test -run x -bench BenchmarkReverseMapping -benchmem -timeout=0 -count 2 -gamma=1.5,2 -keys=long
func TestMain(m *testing.M) {
	var (
		keySizesSlice  = flag.String("keys", "", `list of number of keys to generate (use "long" for 10M, 100M, 1B)`)
		partitionSlice = flag.String("partitions", "", `list of partitions to use (e.g., "{1, 4, 8, 16}")`)
		gammaSlice     = flag.String("gamma", "", `list of gamma values to use (e.g., "{1.0, 1.5, 2.0}")`)
	)
	flag.Parse()

	var err error
	if *keySizesSlice != "" {
		switch *keySizesSlice {
		case "long":
			keySizes = longKeySizes
		default:
			keySizes, err = parseSlice[int](*keySizesSlice)
			check(err)
		}
	}
	if *partitionSlice != "" {
		partitionValues, err = parseSlice[int](*partitionSlice)
		check(err)
	}
	if *gammaSlice != "" {
		gammaValues, err = parseSlice[float64](*gammaSlice)
		check(err)
	}

	for _, size := range keySizes {
		if size >= 5_000_000 && hasTimeout() {
			fmt.Println("Key sizes larger than 5M may cause the test to time out; use the -timeout=0 flag to run longer than 10 minutes")
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// hasTimeout returns true if the test has specified a timeout other than 0.
// This is used to decide whether to run the slow benchmarks or not.
func hasTimeout() bool {
	hasTimeout := true
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "test.timeout" {
			if f.Value.String() == "0s" {
				hasTimeout = false
			}
		}
	})
	return hasTimeout
}

// This is only meant for testing, and should not be used for benchmarking.
type mphf interface{ Find(uint64) uint64 }

type variant[T mphf] struct {
	fn   func(gamma float64, keys []uint64) (T, error)
	fn2  func(gamma float64, partitions int, keys []uint64) (T, error)
	name string
}

func runMPHFTest[T mphf](t *testing.T, tt variant[T], keys []uint64, gamma float64) {
	t.Helper()
	// emit progress every 100k keys
	const progressInterval = 100_000
	size := len(keys)
	logProgress := size > 2*progressInterval
	t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, gamma, size), func(t *testing.T) {
		var bb T
		var err error
		if tt.fn != nil {
			bb, err = tt.fn(gamma, keys)
		} else if tt.fn2 != nil {
			bb, err = tt.fn2(gamma, 20, keys)
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

func getKeymap(keys []uint64, bb *bbhash.BBHash) []uint64 {
	keyMap := make([]uint64, len(keys)+1)
	for _, key := range keys {
		hashIndex := bb.Find(key)
		keyMap[hashIndex] = key
	}
	return keyMap
}

// TestReverseMapping checks that the reverse map returned from NewSequentialWithKeymap is correct.
// First it builds a reverse map the slow way, then it builds a reverse map the fast way.
// Then it compares the two maps.
func TestReverseMapping(t *testing.T) {
	sizes := []uint64{
		1000,
		10_000,
		100_000,
		1_000_000,
	}
	for _, size := range sizes {
		keys := generateKeys(int(size), 99)
		for _, gamma := range []float64{0.5, 1.1, 1.5, 2.0} {
			t.Run(fmt.Sprintf("gamma=%.1f/keys=%d", gamma, size), func(t *testing.T) {
				// Build a reverse map with NewSequential+Find.
				bb, err := bbhash.NewSequential(gamma, keys)
				if err != nil {
					t.Error(err)
				}
				keymap := getKeymap(keys, bb)

				// Build a reverse map with NewSequentialWithKeymap.
				bm, err := bbhash.NewSequentialWithKeymap(gamma, keys)
				if err != nil {
					t.Error(err)
				}

				// Check that the two keymaps are equal.
				for i := range size {
					if bm.Key(i) != keymap[i] {
						// Show only the high 16 bits of the key.
						t.Errorf("bm.Key(%d) = %x, want %x", i, bm.Key(i)>>48, keymap[i]>>48)
					}
				}

				// Check that Key() returns the correct key for the boundary indices.
				tests := []struct {
					index    uint64
					wantZero bool
				}{
					{0, true},
					{1, false},
					{size - 1, false},
					{size, false},
					{size + 1, true},
				}
				for _, test := range tests {
					if got := bm.Key(test.index); (got == 0) != test.wantZero {
						t.Errorf("bm.Key(%d) = %x, want %v", test.index, got>>48, test.wantZero)
					}
				}
			})
		}
	}
}

// BenchmarkReverseMapping benchmarks the speed of building a reverse map.
// The original implementation using NewSequential+Find is very slow;
// with 10_000_000 keys it takes more than 13 hours on a Mac Studio M2 Max 64GB.
// The NewSequentialWithKeymap with 1_000_000_000 keys takes less than 6 minutes.
//
//	go test -run x -bench BenchmarkReverseMapping -benchmem -timeout=0 -count 1 > reverse.txt
func BenchmarkReverseMapping(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			b.Run(fmt.Sprintf("NewSequentialWithKeymap/gamma=%.1f/keys=%d", gamma, size), func(b *testing.B) {
				for range b.N {
					bb, _ = bbhash.NewSequentialWithKeymap(gamma, keys)
				}
			})

			if size > 1_000_000 {
				continue // Skip the NewSequential+Find benchmark for large sizes; it's too slow.
			}
			b.Run(fmt.Sprintf("NewSequential+Find/gamma=%.1f/keys=%d", gamma, size), func(b *testing.B) {
				for range b.N {
					bb, _ = bbhash.NewSequential(gamma, keys)
					keymap = getKeymap(keys, bb)
				}
			})
		}
	}
}

var (
	keymap []uint64
	bb     *bbhash.BBHash
)

// BenchmarkBBHashNew benchmarks the construction of a new BBHash using
// sequential and parallel variants. This will take a long time to run,
// especially if you enable large sizes. Thus, to avoid timeouts, you
// should run this with a -timeout=0 argument.
//
//	go test -run x -bench BenchmarkBBHashNew -benchmem -timeout=0 -count 10 > new.txt
//
// Then compare with:
//
//	benchstat -col /name new.txt
//
// Optionally, you can also compile the test binary and then run it with perf (Linux only):
//
//	go test -c ./
//	perf stat ./bbhash.test -test.run=none -test.bench=BBHashNew -test.timeout=0 -test.count=1
//
// Note that the perf command requires that you have disabled the perf_event_paranoid setting:
//
//	sudo sysctl -w kernel.perf_event_paranoid=0
func BenchmarkBBHashNew(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(fmt.Sprintf("gamma=%.1f/partitions=%3d/keys=%d", gamma, partitions, size), func(b *testing.B) {
					// Note: when partitions is 1, New invokes NewSequential indirectly.
					// This might be slightly slower than using NewSequential directly.
					bb, _ := bbhash.New(gamma, partitions, keys)
					bpk := bb.BitsPerKey()
					b.ResetTimer()
					for range b.N {
						bb, _ = bbhash.New(gamma, partitions, keys)
					}
					// This metric is always the same for a given set of keys.
					b.ReportMetric(bpk, "bits/key")
				})
			}
		}
	}
}

func BenchmarkBBHashFind(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(fmt.Sprintf("gamma=%.1f/partitions=%3d/keys=%d", gamma, partitions, size), func(b *testing.B) {
					bb, _ := bbhash.New(gamma, partitions, keys)
					bpk := bb.BitsPerKey()
					b.ResetTimer()
					for range b.N {
						for _, k := range keys {
							if bb.Find(k) == 0 {
								b.Fatalf("can't find the key: %#x", k)
							}
						}
					}
					// This metric is always the same for a given set of keys.
					b.ReportMetric(bpk, "bits/key")
				})
			}
		}
	}
}

// BenchmarkGammaLevels searches for the gamma value that produces the maximum number of levels.
// This is useful for analyzing the gamma values for varying number of keys, and how it impacts
// the number of bits per key and the number of levels. This can help guide the choice of gamma,
// partitions, and the initial levels for the BBHash.
//
// This benchmark is slow, and should be run with a -timeout=0 argument.
//
//	go test -run x -bench BenchmarkGammaLevels -benchmem -timeout=0 -partitions=4 -gamma=0.5,1,1.5,2.0 -keys=10_000_000,20_000_000 > gamma.txt
func BenchmarkGammaLevels(b *testing.B) {
	// The number of seeds to try for the given number of keys.
	keysToSeeds := func(size int) int {
		if size < 1_000_000_000 {
			return 1_000_000_000 / size
		}
		return 1
	}
	for _, size := range keySizes {
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				b.Run(fmt.Sprintf("gamma=%.1f/partitions=%3d/keys=%d", gamma, partitions, size), func(b *testing.B) {
					keys := generateKeys(size, 99)
					maxBB, _ := bbhash.New(gamma, partitions, keys)
					maxLvl, _ := maxBB.MaxMinLevels()
					maxLvlSeed := 0
					for seed := range keysToSeeds(size) {
						keys := generateKeys(size, seed)
						for range b.N {
							bb, _ := bbhash.New(gamma, partitions, keys)
							lvl, _ := bb.MaxMinLevels()
							if lvl > maxLvl {
								maxBB, maxLvl, maxLvlSeed = bb, lvl, seed
							}
						}
					}
					// Suppress the built-in metric for ns/op.
					b.ReportMetric(0, "ns/op")
					b.ReportMetric(maxBB.BitsPerKey(), "bits/key")
					b.ReportMetric(float64(maxLvl), "max_levels")
					// seed that produced the max levels
					b.ReportMetric(float64(maxLvlSeed), "max_seed")
					// number of seeds tried
					b.ReportMetric(float64(keysToSeeds(size)), "seed_size")
				})
			}
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// parseSlice parses a slice of numbers from a string.
// The input string may be formatted as (with or without spaces):
//   - []int{1, 2, 3}
//   - {1, 2, 3}
//   - [1, 2, 3]
//   - 1, 2, 3
//   - 1.0, 1.5, 2.0
//   - 1000, 100_000, 200_000
func parseSlice[T float64 | int](s string) ([]T, error) {
	slice := make([]T, 0)
	s = strings.TrimPrefix(s, "[]float64")
	s = strings.TrimPrefix(s, "[]float")
	s = strings.TrimPrefix(s, "[]int64")
	s = strings.TrimPrefix(s, "[]int")
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	for _, k := range strings.Split(s, ",") {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		k = strings.ReplaceAll(k, "_", "")
		var i T
		switch any(i).(type) {
		case float64:
			v, err := strconv.ParseFloat(k, 64)
			if err != nil {
				return nil, err
			}
			i = T(v)
		case int:
			v, err := strconv.Atoi(k)
			if err != nil {
				return nil, err
			}
			i = T(v)
		}
		slice = append(slice, i)
	}
	return slice, nil
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
