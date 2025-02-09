package bbhash

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

// This is an interesting benchmark. It shows the performance difference between zeroing two slices
// in a single loop vs zeroing them in two separate loops vs using the clear function vs creating a new slice.
//
// It looks like creating a new slice is faster than clearing a slice when the slice is larger than 1_000_000 elements.
// Need to run the same benchmark on M2 Max in the office and on bbchain and gorina6.
//
// $ benchstat -table "" -col /func -row /size bcvector-bench.txt
//            │ combined_zero │            separate_zero            │                clear                 │             new_alloc_vec              │
//            │    sec/op     │   sec/op     vs base                │    sec/op     vs base                │    sec/op      vs base                 │
// 1000           485.3n ± 1%   351.2n ± 5%  -27.63% (p=0.000 n=10)   211.8n ± 11%  -56.36% (p=0.000 n=10)   3528.0n ±  6%  +626.97% (p=0.000 n=10)
// 10000          5.731µ ± 1%   3.858µ ± 3%  -32.68% (p=0.000 n=10)   2.049µ ± 14%  -64.25% (p=0.000 n=10)   11.062µ ±  2%   +93.02% (p=0.000 n=10)
// 100000         55.40µ ± 2%   40.03µ ± 3%  -27.75% (p=0.000 n=10)   20.18µ ± 11%  -63.58% (p=0.000 n=10)    76.63µ ± 12%   +38.31% (p=0.000 n=10)
// 1000000       1143.8µ ± 3%   404.1µ ± 4%  -64.67% (p=0.000 n=10)   204.0µ ±  5%  -82.16% (p=0.000 n=10)    393.8µ ±  1%   -65.57% (p=0.000 n=10)
// 10000000       9.643m ± 5%   4.110m ± 3%  -57.38% (p=0.000 n=10)   2.017m ±  6%  -79.09% (p=0.000 n=10)    1.470m ±  1%   -84.76% (p=0.000 n=10)
// 100000000     128.44m ± 5%   42.37m ± 3%  -67.01% (p=0.000 n=10)   19.61m ±  4%  -84.74% (p=0.000 n=10)    13.27m ±  3%   -89.67% (p=0.000 n=10)
// 1000000000      3.094 ± 2%    2.907 ± 4%   -6.03% (p=0.000 n=10)    2.620 ±  3%  -15.32% (p=0.000 n=10)     2.806 ± 13%    -9.32% (p=0.000 n=10)
// geomean        945.5µ        525.7µ       -44.40%                  292.9µ        -69.02%                   675.9µ         -28.52%

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
		vsink := newBCVector(size)
		b.Run(fmt.Sprintf("combined_zero/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				vsink.combined_zero()
			}
		})
		runtime.GC()
		b.Run(fmt.Sprintf("separate_zero/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				vsink.separate_zero()
			}
		})
		runtime.GC()
		b.Run(fmt.Sprintf("clear/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				vsink.clear()
			}
		})
		runtime.GC()
	}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("new_alloc_vec/size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				newBCVector(size)
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

func (b *bcVector) clear() {
	clear(b.c)
	clear(b.v)
}
