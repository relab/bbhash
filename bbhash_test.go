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
	"github.com/relab/bbhash/internal/test"
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
//	go test -run x -bench BenchmarkReverseMapping -benchmem -timeout=0 -count 2 -gamma=1.5,2 -partitions=1,2,4,8 -keys=long
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
	for _, g := range gammaValues {
		t.Run(test.Name("StarWars", []string{"gamma", "keys"}, g, len(keys)), func(t *testing.T) {
			bb, err := bbhash.New(keys, bbhash.Gamma(g))
			if err != nil {
				t.Fatal(err)
			}
			validateKeyMappings(t, bb, keys)
		})
	}
}

func TestManyKeys(t *testing.T) {
	sizes := []int{
		1000,
		10_000,
		100_000,
	}
	const seed = 123
	tcs := []struct {
		name string
		opts []bbhash.Options
	}{
		{name: "ReverseMap", opts: []bbhash.Options{bbhash.WithReverseMap()}},
		{name: "Parallel", opts: []bbhash.Options{bbhash.Parallel()}},
		{name: "Partitioned4", opts: []bbhash.Options{bbhash.Partitions(4)}},
		{name: "Partitioned8", opts: []bbhash.Options{bbhash.Partitions(8)}},
		{name: "Partitioned15", opts: []bbhash.Options{bbhash.Partitions(15)}},
	}
	for _, tc := range tcs {
		for _, gamma := range []float64{1.1, 1.5, 2.0, 2.5, 3.0, 5.0} {
			for _, size := range sizes {
				t.Run(test.Name(tc.name, []string{"gamma", "keys"}, gamma, size), func(t *testing.T) {
					keys := generateKeys(size, seed)
					bb, err := bbhash.New(keys, append(tc.opts, bbhash.Gamma(gamma))...)
					if err != nil {
						t.Fatal(err)
					}
					validateKeyMappings(t, bb, keys)
				})
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
	tcs := []struct {
		name string
		opts []bbhash.Options
	}{
		{name: "Partitioned4", opts: []bbhash.Options{bbhash.Partitions(4)}},
	}
	const gamma = 2.0
	for _, tc := range tcs {
		for _, size := range sizes {
			t.Run(test.Name(tc.name, []string{"gamma", "keys"}, gamma, size), func(t *testing.T) {
				keys := generateKeys(size, 99)
				bb, err := bbhash.New(keys, append(tc.opts, bbhash.Gamma(gamma))...)
				if err != nil {
					t.Fatal(err)
				}
				validateKeyMappings(t, bb, keys)
			})
		}
	}
}

// getReverseMap returns a reverse map from indices to keys.
// The reverse map is built by calling Find for each key.
// This is used to compare the reverse map built by WithReverseMap option.
func getReverseMap(keys []uint64, bb mphf) []uint64 {
	keyMap := make([]uint64, len(keys)+1)
	for _, key := range keys {
		hashIndex := bb.Find(key)
		keyMap[hashIndex] = key
	}
	return keyMap
}

// TestReverseMapping checks that the reverse map returned from New(WithReverseMap) is correct.
// First it builds a reverse map the slow way, then it builds a reverse map the fast way.
// Then it compares the two maps.
func TestReverseMapping(t *testing.T) {
	sizes := []uint64{
		1000,
		10_000,
		100_000,
	}
	for _, size := range sizes {
		keys := generateKeys(int(size), 99)
		for _, gamma := range []float64{0.5, 1.5, 2.0} {
			for _, partitions := range []int{1, 2, 3, 5, 20} {
				t.Run(test.Name("Params", []string{"gamma", "partitions", "keys"}, gamma, partitions, size), func(t *testing.T) {
					bb, err := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
					if err != nil {
						t.Error(err)
					}
					// Build the reverse map using Find.
					reverseMap := getReverseMap(keys, bb)

					// Build the reverse map directly using WithReverseMap option.
					bm, err := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions), bbhash.WithReverseMap())
					if err != nil {
						t.Error(err)
					}

					// Check that the two reverse maps are equal.
					for i := range size {
						if bm.Key(i) != reverseMap[i] {
							// Show only the high 16 bits of the key.
							t.Errorf("bm.Key(%d) = %x, want %x", i, bm.Key(i)>>48, reverseMap[i]>>48)
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
}

// BenchmarkReverseMapping benchmarks the speed of building a reverse map.
// The original implementation using New(Sequential)+Find is very slow;
// with 10_000_000 keys it takes more than 13 hours on a Mac Studio M2 Max 64GB.
// The New(WithReverseMap) with 1_000_000_000 keys takes less than 6 minutes.
//
//	go test -run x -bench BenchmarkReverseMapping -benchmem -timeout=0 -count 1 > reverse.txt
func BenchmarkReverseMapping(b *testing.B) {
	for _, size := range keySizes {
		keys := generateKeys(size, 99)
		for _, gamma := range gammaValues {
			for _, partitions := range partitionValues {
				name := test.Name("New(WithReverseMap)", []string{"gamma", "partitions", "keys"}, gamma, partitions, size)
				b.Run(name, func(b *testing.B) {
					for b.Loop() {
						bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions), bbhash.WithReverseMap())
					}
				})

				if size > 1_000_000 {
					continue // Skip the New(Sequential)+Find benchmark for large sizes; it's too slow.
				}
				name = test.Name("New(Sequential)+Find", []string{"gamma", "partitions", "keys"}, gamma, partitions, size)
				b.Run(name, func(b *testing.B) {
					for b.Loop() {
						bb, _ := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
						getReverseMap(keys, bb)
					}
				})
			}
		}
	}
}

// BenchmarkBBHashNew benchmarks the construction of a new BBHash using
// sequential and partition variants. This will take a long time to run,
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
				name := test.Name("Params", []string{"gamma", "partitions", "keys"}, gamma, partitions, size)
				b.Run(name, func(b *testing.B) {
					bb, _ := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
					bpk := bb.BitsPerKey()
					b.ResetTimer()
					for b.Loop() {
						bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
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
				name := test.Name("Params", []string{"gamma", "partitions", "keys"}, gamma, partitions, size)
				b.Run(name, func(b *testing.B) {
					bb, _ := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
					bpk := bb.BitsPerKey()
					b.ResetTimer()
					for b.Loop() {
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
				name := test.Name("Params", []string{"gamma", "partitions", "keys"}, gamma, partitions, size)
				b.Run(name, func(b *testing.B) {
					keys := generateKeys(size, 99)
					maxBB, _ := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
					maxLvl, _ := maxBB.MaxMinLevels()
					maxLvlSeed := 0
					for seed := range keysToSeeds(size) {
						keys := generateKeys(size, seed)
						for b.Loop() {
							bb, _ := bbhash.New(keys, bbhash.Gamma(gamma), bbhash.Partitions(partitions))
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

// mphf provides an interface to find keys in a minimal perfect hash function.
// This is only meant for testing, and should not be used for benchmarking.
type mphf interface{ Find(uint64) uint64 }

// validateKeyMappings checks that the keys are correctly mapped to the indices.
// It also checks that the indices are unique.
// This check can be slow for large key sets.
func validateKeyMappings(t *testing.T, bb mphf, keys []uint64) {
	t.Helper()

	const progressInterval = 5 * time.Second
	nextLogTime := time.Now().Add(progressInterval)

	entries := uint64(len(keys))
	keyMap := make(map[uint64]uint64, entries)
	for keyIndex, key := range keys {
		if time.Now().After(nextLogTime) {
			t.Logf("%d keys checked so far", keyIndex)
			nextLogTime = time.Now().Add(progressInterval)
		}

		hashIndex := bb.Find(key)
		if hashIndex == 0 {
			t.Fatalf("can't find key: %#x", key)
		}
		if hashIndex > entries {
			t.Fatalf("key %d <%#x> mapping %d out-of-bounds", keyIndex, key, hashIndex)
		}
		if x, ok := keyMap[hashIndex]; ok {
			t.Errorf("index %d already mapped to key %#x", hashIndex, x)
		}
		keyMap[hashIndex] = key
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
