package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/relab/bbhash"
)

func main() {
	var (
		name       = flag.String("name", "seq", "name of the mphf to benchmark (seq, par, par2)")
		gamma      = flag.Float64("gamma", 2.0, "gamma parameter")
		salt       = flag.Uint64("salt", 99, "salt parameter")
		keys       = flag.Int("keys", -1, "number of keys to use (default: runs over a range of keys; can take a long time)")
		partitions = flag.Int("partitions", 1, "number of partitions to use (for par2 only)")
		count      = flag.Int("count", 1, "number of times to run the benchmark")
	)
	flag.Parse()

	perKeyElapsed := make(map[int][]time.Duration, *count)
	perKeyElapsedFind := make(map[int][]time.Duration, *count)
	levels := make(map[int]int, *count)
	bitsPerKey := make(map[int]float64, *count)
	for _, numKeys := range keyRange(keys) {
		elapsed, elapsedFind, lvls, bpk := run(*name, numKeys, *gamma, *salt, *count, *partitions)
		perKeyElapsed[numKeys] = elapsed
		perKeyElapsedFind[numKeys] = elapsedFind
		levels[numKeys] = lvls
		bitsPerKey[numKeys] = bpk
	}
	filename := fmt.Sprintf("bbhash-%s-gamma-%.1f-partitions-%d.csv", *name, *gamma, *partitions)
	writeCSVFile(filename, perKeyElapsed, perKeyElapsedFind, levels, bitsPerKey)
}

func keyRange(keys *int) []int {
	if *keys == -1 {
		// return []int{1000, 10_000, 100_000}
		return []int{1000, 10_000, 100_000, 1000_000, 10_000_000, 100_000_000}
	}
	return []int{*keys}
}

func run(name string, keys int, gamma float64, salt uint64, count, partitions int) ([]time.Duration, []time.Duration, int, float64) {
	switch name {
	case "seq":
		return runSequential(keys, gamma, salt, count)
	case "par":
		return runParallel(keys, gamma, salt, count)
	case "par2":
		return runParallel2(keys, partitions, gamma, salt, count)
	default:
		panic("unknown mphf name")
	}
}

func writeCSVFile(filename string, create, find map[int][]time.Duration, levels map[int]int, bpk map[int]float64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	sortedKeys := make([]int, 0, len(create))
	for keys := range create {
		sortedKeys = append(sortedKeys, keys)
	}
	sort.Ints(sortedKeys)

	// Write header
	if err := w.Write([]string{"Keys", "Levels", "BitsPerKey", "CreateTime", "FindTime"}); err != nil {
		return err
	}
	for _, keys := range sortedKeys {
		createElapsed := create[keys]
		findElapsed := find[keys]
		lvls := strconv.Itoa(levels[keys])
		bitsPerKey := fmt.Sprintf("%.1f", bpk[keys])
		for i := range createElapsed {
			createTime := float64(createElapsed[i]) / float64(time.Millisecond)
			findTime := float64(findElapsed[i]) / float64(time.Millisecond)
			c := fmt.Sprintf("%.4f", createTime)
			f := fmt.Sprintf("%.4f", findTime)
			if err := w.Write([]string{strconv.Itoa(keys), lvls, bitsPerKey, c, f}); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func runSequential(numKeys int, gamma float64, salt uint64, count int) ([]time.Duration, []time.Duration, int, float64) {
	keys := generateKeys(numKeys, 99)
	var bb *bbhash.BBHash
	var err error
	elapsed := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		start := time.Now()
		bb, err = bbhash.NewSequential(gamma, salt, keys)
		elapsed[i] = time.Since(start)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Sequential:", bb)
	return elapsed, findAll(bb, keys, count), bb.Levels(), bb.BitsPerKey()
}

func runParallel(numKeys int, gamma float64, salt uint64, count int) ([]time.Duration, []time.Duration, int, float64) {
	keys := generateKeys(numKeys, 99)
	var bb *bbhash.BBHash
	var err error
	elapsed := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		start := time.Now()
		bb, err = bbhash.NewParallel(gamma, salt, keys)
		elapsed[i] = time.Since(start)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Parallel:", bb)
	return elapsed, findAll(bb, keys, count), bb.Levels(), bb.BitsPerKey()
}

func runParallel2(numKeys, numPartitions int, gamma float64, salt uint64, count int) ([]time.Duration, []time.Duration, int, float64) {
	keys := generateKeys(numKeys, 99)
	var bb *bbhash.BBHash2
	var err error
	elapsed := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		start := time.Now()
		bb, err = bbhash.NewParallel2(gamma, numPartitions, salt, keys)
		elapsed[i] = time.Since(start)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Parallel2:", bb)
	// We return only the max level for now
	max, _ := bb.MaxMinLevels()
	return elapsed, findAll(bb, keys, count), max, bb.BitsPerKey()
}

func findAll(bb interface{ Find(uint64) uint64 }, keys []uint64, count int) []time.Duration {
	elapsed := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		start := time.Now()
		for _, key := range keys {
			index := bb.Find(key)
			if index == 0 {
				panic("key not found")
			}
		}
		elapsed[i] = time.Since(start)
	}
	return elapsed
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
