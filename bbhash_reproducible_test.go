package bbhash_test

import (
	"fmt"
	"math/rand"
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
		name  string
		gamma float64
		seed  int
		fn    func(gamma float64, salt uint64, keys []uint64) (*bbhash.BBHash, error)
	}{
		{name: "Sequential", gamma: 2.0, seed: 123, fn: bbhash.NewSequential},
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
		0x8141020141023024,
		0x06cb2d0060086cb0,
		0x020019806ec20294,
		0x30444815086a4100,
		0x10954a0025044a88,
		0x9704081057586080,
		0x8063651823c92457,
		0x64511101570263a8,
		0xe82978a62206a246,
		0x2000b59c0c3d0013,
		0xc2008021560d900b,
		0xa3809190011a770f,
		0x0d8413428a800040,
		0x454006511028028f,
		0x31214869dd002570,
		0x28441d8031040800,
		0x0520211c808ae280,
		0x0a09285720480201,
		0x242021e816c8d174,
		0x14032006049910c1,
		0x4d228450059151c6,
		0x00a518210ced508c,
		0x0677285c0c54103c,
		0x20804780010424a2,
		0x0608d0010e687256,
		0x43e1486345aa1181,
		0x84854a8ae0403236,
		0x2444a40c03024902,
		0x82080c3504006c20,
		0xd104b4d702c00e64,
		0x4310b2c610102382,
		0x0b31a81102300101,
	},
	// Level 1:
	{
		0x0b43250338076052,
		0x12104c4a8079408b,
		0x5115120541902c00,
		0x4a8400d002509068,
		0xa010ba4118440a04,
		0x088024801b529000,
		0x011520a389a4e600,
		0x00482e8039ad8200,
		0x00c1704118b540ed,
		0x8a00800fb22a14c2,
		0xa00f8b395fa411c8,
		0x7a02c6340100c240,
	},
	// Level 2:
	{
		0x248863d1288e40d0,
		0x8002808493168156,
		0x8281230b4003430e,
		0xac84754020500282,
		0x3f1088f024240012,
	},
	// Level 3:
	{
		0x020c06cce2007310,
		0x13a0202500084c31,
	},
	// Level 4:
	{
		0x01e1140011a00004,
	},
	// Level 5:
	{
		0x0200800009204808,
	},
}
