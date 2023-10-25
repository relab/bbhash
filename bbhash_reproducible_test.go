package bbhash_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/relab/bbhash"
)

// Run this test as follows to print the bit vectors:
//
//	% PRINT=1 go test -run TestReproducibleBitVectors > reproducible.txt
//
// Then copy the output into this file, update the entries below, and run:
//
//	% go test -run TestReproducibleBitVectors
func TestReproducibleBitVectors(t *testing.T) {
	printVector := os.Getenv("PRINT") != ""

	sizes := []int{
		1000,
	}
	tests := []struct {
		fn    func(gamma float64, keys []uint64) (*bbhash.BBHash, error)
		name  string
		gamma float64
		seed  int
	}{
		{name: "Sequential", gamma: 2.0, seed: 123, fn: bbhash.NewSequential},
		{name: "Parallel__", gamma: 2.0, seed: 123, fn: bbhash.NewParallel},
	}

	for _, tt := range tests {
		for _, size := range sizes {
			keys := generateKeys(size, tt.seed)
			t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, tt.gamma, size), func(t *testing.T) {
				bb, err := tt.fn(tt.gamma, keys)
				if err != nil {
					t.Fatal(err)
				}
				if printVector {
					fmt.Println(bb.BitVectors())
				}
				want := bitVectors1000
				got := bb.LevelVectors()
				if diff := diff(want, got); diff != "" {
					t.Errorf("bit vectors mismatch (-want +got):\n%s", diff)
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
}

// diff returns a custom diff of two slices of slices of uint64;
// -a is the wanted value, +b is the got value.
func diff(a, b [][]uint64) string {
	var s strings.Builder
	minLevels := min(len(a), len(b))
	if len(a) != len(b) {
		s.WriteString(fmt.Sprintf("levels: a=%d, b=%d, min=%d\n", len(a), len(b), minLevels))
	}
	for i := 0; i < minLevels; i++ {
		minEntries := min(len(a[i]), len(b[i]))
		if len(a[i]) != len(b[i]) {
			s.WriteString(fmt.Sprintf("entries: a[%d]=%d, b[%d]=%d, min=%d\n", i, len(a[i]), i, len(b[i]), minEntries))
		}
		for j := 0; j < minEntries; j++ {
			w := a[i][j]
			g := b[i][j]
			if w != g {
				xor := w ^ g
				s.WriteString(fmt.Sprintf("-a[%02d,%02d]: %#016x xor: %064b\n", i, j, w, xor))
				s.WriteString(fmt.Sprintf("+b[%02d,%02d]: %#016x xor: %064b\n", i, j, g, xor))
			}
		}
	}
	return s.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var bitVectors1000 = [][]uint64{
	// Level 0:
	{
		0x684050814eee4400,
		0x006e6010b8496202,
		0x28c4e10011005100,
		0x28111a200e105028,
		0xf52310206468484b,
		0x803116c4d124150d,
		0xd45033481c388281,
		0xf2a0921408022a31,
		0x0d0931a809cd1555,
		0x0141c84022b12040,
		0x14e371f8121c1e20,
		0x1d87a21cd0102001,
		0x426c210022013144,
		0x2886230646218210,
		0x0880428b10018180,
		0x0062002790035441,
		0x4087068c0eece0d3,
		0x0940307807c8c0a4,
		0x02f4442440f81024,
		0x0505000b14664021,
		0x2805041e52180e3a,
		0x38c32b40004306d8,
		0x04b4060ee0038204,
		0x82018d032c044d17,
		0x8500450270684884,
		0x80302400008a0080,
		0x1c983419ba380de1,
		0x4002133085155800,
		0xcd00818319830c67,
		0x43111706e4b25236,
		0x2004008018247090,
		0x1740250ca345a4c4,
	},
	// Level 1:
	{
		0xd145040494c48888,
		0xb22810b685214adb,
		0x0050107340804208,
		0x5c23284d084a6008,
		0xb2290420528670c1,
		0x024485460ce00030,
		0x802818403024544a,
		0xa1133ebcc344980c,
		0xb401ab0b3304659d,
		0x8040e624d8329a0c,
		0x2212323a71450080,
		0x6a680840a03422a0,
	},
	// Level 2:
	{
		0xa05b716461b8c004,
		0x411972820f460030,
		0xf406c04d81024900,
		0x59100242805019a0,
		0x8081d00099482102,
	},
	// Level 3:
	{
		0x6101100201001430,
		0x0211e20010604410,
	},
	// Level 4:
	{
		0x0100860400161604,
	},
	// Level 5:
	{
		0x0002440212000000,
	},
}
