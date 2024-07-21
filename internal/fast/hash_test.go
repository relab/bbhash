package fast_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/relab/bbhash/internal/fast"
)

func TestFastHash(t *testing.T) {
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := fast.LevelHash(lvl)
		for key := uint64(0); key < 5; key++ {
			slowHash := fast.Hash(lvl, key)
			fastHash := fast.KeyHash(lvlHash, key)
			if slowHash != fastHash {
				t.Errorf("hash(%d, %d) != keyHash(%#x, %d)", lvl, key, lvlHash, key)
				t.Logf("   hash(lvl=%d,key=%d): %#x", lvl, key, slowHash)
				t.Logf("keyHash(lvl=%d,key=%d): %#x", lvl, key, fastHash)
			}
		}
	}
}

var sink uint64

func BenchmarkHashLevel(b *testing.B) {
	if os.Getenv("HASH") == "" {
		b.Skip("Skipping benchmark, set HASH=1 to run it.")
	}
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := fast.LevelHash(lvl)
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = fast.KeyHash(lvlHash, uint64(key))
			}
		})
	}
}

func BenchmarkHashFull(b *testing.B) {
	if os.Getenv("HASH") == "" {
		b.Skip("Skipping benchmark, set HASH=1 to run it.")
	}
	for lvl := uint64(0); lvl < 5; lvl++ {
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = fast.Hash(lvl, uint64(key))
			}
		})
	}
}
