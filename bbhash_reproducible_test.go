package bbhash_test

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"strings"
	"testing"

	"github.com/relab/bbhash"
)

var update = flag.Bool("update", false, "update bit vectors golden test file")

// To update the bit vectors golden test file, run:
//
//	% go test -run TestReproducibleBitVectors -update
func TestReproducibleBitVectors(t *testing.T) {
	sizes := []int{
		1000,
		10000,
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

	var bitVectors strings.Builder
	var wantBitVectorsFunc strings.Builder
	wantBitVectorsFunc.WriteString("package bbhash_test\n\nfunc wantBitVectors(size int) [][]uint64 {\nswitch size {\n")
	done := make(map[int]bool)
	for _, tt := range tests {
		for _, size := range sizes {
			keys := generateKeys(size, tt.seed)
			t.Run(fmt.Sprintf("name=%s/gamma=%0.1f/keys=%d", tt.name, tt.gamma, size), func(t *testing.T) {
				bb, err := tt.fn(tt.gamma, keys)
				if err != nil {
					t.Fatal(err)
				}
				if *update && !done[size] {
					wantBitVectorsFunc.WriteString(fmt.Sprintf("case %d:\n", size))
					wantBitVectorsFunc.WriteString(fmt.Sprintf("return bitVectors%d\n", size))
					bitVectors.WriteString(bb.BitVectors(size))
					bitVectors.WriteString("\n")
					done[size] = true
				}

				want := wantBitVectors(size)
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
	if *update {
		testFile := "bbhash_reproducible_bit_vectors_test.go"
		t.Logf("rewriting %s", testFile)
		wantBitVectorsFunc.WriteString("}\n	return nil\n  }\n\n")
		s, err := format.Source([]byte(wantBitVectorsFunc.String() + bitVectors.String()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(testFile, s, 0o666); err != nil {
			t.Fatal(err)
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
