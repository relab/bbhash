package bbhash

import (
	"golang.org/x/sync/errgroup"
)

// BBHash2 represents a minimal perfect hash for a set of keys.
type BBHash2 struct {
	partitions []BBHash
	offsets    []int
}

// New creates a new BBHash2 for the given keys. The keys must be unique.
// Creation is configured using the provided options. The default options
// are used if none are provided. Available options include: Gamma,
// InitialLevels, Partitions, Parallel, and WithReverseMap.
// With fewer than 1000 keys, the sequential version is always used.
func New(keys []uint64, opts ...Options) (*BBHash2, error) {
	if len(keys) < 1 {
		panic("bbhash: no keys provided")
	}

	o := newOptions(opts...)
	if o.partitions > 1 && o.parallel {
		panic("bbhash: parallel and partitions not supported")
	}
	if len(keys) < 1000 || o.partitions == 1 {
		bb := newBBHash(o.initialLevels)
		var err error
		switch {
		case !o.reverseMap && !o.parallel:
			err = bb.compute(keys, o.gamma)
		case o.reverseMap && !o.parallel:
			err = bb.computeWithKeymap(keys, o.gamma)
		case !o.reverseMap && o.parallel:
			err = bb.computeParallel(keys, o.gamma)
		case o.reverseMap && o.parallel:
			panic("bbhash: parallel and reverse map not supported")
		}
		if err != nil {
			return nil, err
		}
		return &BBHash2{
			partitions: []BBHash{bb},
			offsets:    []int{0},
		}, nil
	}
	return newPartitioned(keys, o)
}

// newPartitioned partitions the keys and creates multiple BBHashes in parallel.
func newPartitioned(keys []uint64, o *options) (*BBHash2, error) {
	// Partition the keys into partitions by placing keys with the
	// same remainder (modulo partitions) into the same partition.
	// This approach copies the keys into partitions slices, which
	// may lead to some variation in the number of keys in each partition.
	partitionKeys := make([][]uint64, o.partitions)
	for _, k := range keys {
		i := k % uint64(o.partitions)
		partitionKeys[i] = append(partitionKeys[i], k)
	}
	bb := &BBHash2{
		partitions: make([]BBHash, o.partitions),
		offsets:    make([]int, o.partitions),
	}
	grp := &errgroup.Group{}
	for offset, j := 0, 0; j < o.partitions; j++ {
		bb.offsets[j] = offset
		offset += len(partitionKeys[j])
		grp.Go(func() error {
			bb.partitions[j] = newBBHash(o.initialLevels)
			if o.reverseMap {
				return bb.partitions[j].computeWithKeymap(partitionKeys[j], o.gamma)
			}
			return bb.partitions[j].compute(partitionKeys[j], o.gamma)
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
func (bb BBHash2) Find(key uint64) uint64 {
	i := key % uint64(len(bb.partitions))
	return bb.partitions[i].Find(key) + uint64(bb.offsets[i])
}

// Key returns the key for the given index.
// The index must be in the range [1, len(keys)], otherwise 0 is returned.
func (bb BBHash2) Key(index uint64) uint64 {
	for _, b := range bb.partitions {
		if index < uint64(len(b.reverseMap)) {
			return b.reverseMap[index]
		}
		index -= uint64(len(b.reverseMap)) - 1
	}
	return 0
}

// Partitions returns the number of partitions in the BBHash2.
// This is mainly useful for testing and may be removed in the future.
func (bb BBHash2) Partitions() int {
	return len(bb.partitions)
}

// SinglePartition returns the underlying BBHash if it contains a single partition.
// If there are multiple partitions, it returns nil.
func (bb BBHash2) SinglePartition() *BBHash {
	if len(bb.partitions) == 1 {
		return &bb.partitions[0]
	}
	return nil
}

// enforce interface compliance
var (
	_ bbhash     = (*BBHash2)(nil)
	_ reverseMap = (*BBHash2)(nil)
)
