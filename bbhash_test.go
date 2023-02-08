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
	bb, err := bbhash.NewSerial(2.0, salt, keys)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(bb)
	keyMap := make(map[uint64]uint64)
	for keyIndex, key := range keys {
		hashIndex := bb.Find(key)
		checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)

		if x, ok := keyMap[hashIndex]; ok {
			t.Fatalf("index %d already mapped to key %#x", hashIndex, x)
		}
		keyMap[hashIndex] = key
	}
}

func TestManyKeys(t *testing.T) {
	tests := []struct {
		gamma float64
		size  int
	}{
		{gamma: 2.0, size: 1000},
		{gamma: 2.0, size: 10000},
		{gamma: 2.0, size: 100000},
		{gamma: 2.0, size: 1000000},
		{gamma: 2.0, size: 10000000},
	}
	// Find() is too slow to check 10 million keys; the test will potentially run for 15-20 minutes.
	const expectedTimeToFind10MillionKeys = 18 * time.Minute
	deadline, ok := t.Deadline()

	salt := rand.New(rand.NewSource(99)).Uint64()
	for _, tt := range tests {
		keys := generateKeys(tt.size, 99)
		t.Run(fmt.Sprintf("gamma=%.1f/keys=%d", tt.gamma, tt.size), func(t *testing.T) {
			bb, err := bbhash.NewSerial(tt.gamma, salt, keys)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(bb)
			// Skip test key sizes that are too large for the test to finish in time.
			// If ok == false, the test has no deadline (timeout=0), in which case we don't skip the test.
			if tt.size > 1_000_000 && ok && time.Until(deadline) < expectedTimeToFind10MillionKeys {
				t.Skip("test will timeout before it can finish; use -timeout=0 or -timeout=20m to run it anyway")
			}
			for keyIndex, key := range keys {
				hashIndex := bb.Find(key)
				checkKey(t, keyIndex, key, uint64(len(keys)), hashIndex)
			}
		})
	}
}

var bbSink *bbhash.BBHash

func BenchmarkNewSerial(b *testing.B) {
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
				bbSink, _ = bbhash.NewSerial(tt.gamma, salt, keys)
			}
		})
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
		bb, err := bbhash.NewSerial(tt.gamma, salt, keys)
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
