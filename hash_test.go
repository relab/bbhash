package bbhash

import (
	"fmt"
	"testing"
)

func TestFastHash(t *testing.T) {
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := levelHash(lvl)
		for key := uint64(0); key < 5; key++ {
			slowHash := hash(lvl, key)
			fastHash := keyHash(lvlHash, key)
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
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := levelHash(lvl)
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = keyHash(lvlHash, uint64(key))
			}
		})
	}
}

func BenchmarkHashFull(b *testing.B) {
	for lvl := uint64(0); lvl < 5; lvl++ {
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = hash(lvl, uint64(key))
			}
		})
	}
}
