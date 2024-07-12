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
// For small key sets, you may want to use NewSequential instead, since it will likely
// be faster. NewParallel2 allocates more memory than NewSequential, but will be faster
// for large key sets.
//
// This partitions the input and creates multiple BBHashes using multiple goroutines.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func NewParallel2(gamma float64, numPartitions int, keys []uint64) (*BBHash2, error) {
	gamma = max(gamma, minimalGamma)
	// Partition the keys into numPartitions by placing keys with the
	// same remainder (modulo numPartitions) into the same partition.
	// This approach copies the keys into numPartitions slices, which
	// may lead to some variation in the number of keys in each partition.
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

// Find returns a unique index representing the key in the minimal hash set.
//
// The return value is meaningful ONLY for keys in the original key set
// (provided at the time of construction of the minimal hash set).
//
// If the key is in the original key set, the return value is guaranteed to be
// in the range [1, len(keys)].
//
// If the key is not in the original key set, two things can happen:
// 1. The return value is 0, representing that the key was not in the original key set.
// 2. The return value is in the expected range [1, len(keys)], but is a false positive.
func (bb *BBHash2) Find(key uint64) uint64 {
	i := key % uint64(len(bb.partitions))
	return bb.partitions[i].Find(key) + uint64(bb.offsets[i])
}
