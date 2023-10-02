package bbhash

import (
	"golang.org/x/sync/errgroup"
)

type BBHash2 struct {
	partitions []*BBHash
	offsets    []int
}

// New creates a new BBHash2 for the given keys. The keys must be unique.
// If partitions is 1 or less, then a single BBHash is created, wrapped in a BBHash2.
// Otherwise, the keys are partitioned into the given the number partitions,
// and multiple BBHashes are created in parallel.
//
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func New(gamma float64, partitions int, keys []uint64) (*BBHash2, error) {
	if partitions <= 1 {
		bb, err := NewSequential(gamma, keys)
		if err != nil {
			return nil, err
		}
		return &BBHash2{
			partitions: []*BBHash{bb},
			offsets:    []int{0},
		}, nil
	}
	return NewParallel2(gamma, partitions, keys)
}

// NewParallel2 creates a new BBHash2 for the given keys. The keys must be unique.
// This partitions the input and creates multiple BBHashes using multiple goroutines.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func NewParallel2(gamma float64, numPartitions int, keys []uint64) (*BBHash2, error) {
	partitionKeys := make([][]uint64, numPartitions)
	for _, k := range keys {
		i := k % uint64(numPartitions)
		partitionKeys[i] = append(partitionKeys[i], k)
	}
	bb := &BBHash2{
		partitions: make([]*BBHash, numPartitions),
		offsets:    make([]int, numPartitions),
	}
	grp := &errgroup.Group{}
	for offset, j := 0, 0; j < numPartitions; j++ {
		j := j
		bb.offsets[j] = offset
		offset += len(partitionKeys[j])
		grp.Go(func() error {
			bb.partitions[j] = newBBHash()
			return bb.partitions[j].compute(gamma, partitionKeys[j])
		})
	}
	if err := grp.Wait(); err != nil {
		return nil, err
	}
	return bb, nil
}

func (bb *BBHash2) Find(key uint64) uint64 {
	i := key % uint64(len(bb.partitions))
	return bb.partitions[i].Find(key) + uint64(bb.offsets[i])
}
