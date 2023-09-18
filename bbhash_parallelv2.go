package bbhash

import (
	"golang.org/x/sync/errgroup"
)

type BBHash2 struct {
	partitions []*BBHash
	offsets    []int
}

// NewParallel2 creates a new BBHash for the given keys. The keys must be unique.
// This partitions the input and creates multiple BBHashes using multiple goroutines.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more memory will be consumed by the BBHash.
func NewParallel2(gamma float64, partitionSize int, keys []uint64) (*BBHash2, error) {
	partitionKeys := make([][]uint64, partitionSize)
	for _, k := range keys {
		i := k % uint64(partitionSize)
		partitionKeys[i] = append(partitionKeys[i], k)
	}
	bb := &BBHash2{
		partitions: make([]*BBHash, partitionSize),
		offsets:    make([]int, partitionSize),
	}
	grp := &errgroup.Group{}
	for offset, j := 0, 0; j < partitionSize; j++ {
		j := j
		bb.offsets[j] = offset
		offset += len(partitionKeys[j])
		grp.Go(func() error {
			bb.partitions[j] = newBBHash()
			return bb.partitions[j].compute(partitionKeys[j], gamma)
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
