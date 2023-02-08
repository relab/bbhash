package bbhash

import (
	"fmt"
	"testing"
)

func TestFastHash(t *testing.T) {
	salt := uint64(0xdeadbeef)
	saltHash := saltHash(salt)
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := levelHash(saltHash, lvl)
		for key := uint64(0); key < 5; key++ {
			slowHash := hash(saltHash, lvl, key)
			fastHash := keyHash(lvlHash, key)
			if slowHash != fastHash {
				t.Errorf("hash(%#x, %d, %d) != keyHash(%#x, %d)", saltHash, lvl, key, lvlHash, key)
				t.Logf("   hash(saltHash=%#x,lvl=%d,key=%d): %#x", saltHash, lvl, key, slowHash)
				t.Logf("keyHash(saltHash=%#x,lvl=%d,key=%d): %#x", saltHash, lvl, key, fastHash)
			}
		}
	}
}

var sink uint64

func BenchmarkHashLevel(b *testing.B) {
	salt := uint64(0xdeadbeef)
	saltHash := saltHash(salt)
	for lvl := uint64(0); lvl < 5; lvl++ {
		lvlHash := levelHash(saltHash, lvl)
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = keyHash(lvlHash, uint64(key))
			}
		})
	}
}

func BenchmarkHashFull(b *testing.B) {
	salt := uint64(0xdeadbeef)
	saltHash := saltHash(salt)
	for lvl := uint64(0); lvl < 5; lvl++ {
		b.Run(fmt.Sprintf("lvl=%d", lvl), func(b *testing.B) {
			for key := 0; key < b.N; key++ {
				sink = hash(saltHash, lvl, uint64(key))
			}
		})
	}
}
