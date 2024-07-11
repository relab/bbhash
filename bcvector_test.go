package bbhash

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

var vsink *bcVector

func BenchmarkBCVectorZero(b *testing.B) {
	if os.Getenv("ZERO_VECTOR") == "" {
		b.Skip("Skipping benchmark, set ZERO_VECTOR=1 to run it.")
	}

	sizes := []uint64{
		1000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
		1_000_000_000,
	}
	for _, size := range sizes {
		vsink = newBCVector(size)
		b.Run(fmt.Sprintf("combined_zero/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				vsink.combined_zero()
			}
		})
		runtime.GC()
		b.Run(fmt.Sprintf("separate_zero/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				vsink.separate_zero()
			}
		})
		runtime.GC()
	}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("new_alloc_vec/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				vsink = newBCVector(size)
			}
		})
	}
}

func (b *bcVector) combined_zero() {
	for i := range b.c {
		b.c[i] = 0
		b.v[i] = 0
	}
}

func (b *bcVector) separate_zero() {
	for i := range b.c {
		b.c[i] = 0
	}
	for i := range b.c {
		b.v[i] = 0
	}
}
