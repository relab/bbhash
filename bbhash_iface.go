package bbhash

// These interfaces are unexported and are only used to enforce the interface compliance.

// bbhash is an interface for the lookup operation of the BBHash.
type bbhash interface {
	// Find returns the index of the key in the BBHash.
	Find(key uint64) uint64
}

// reverseMap is an interface for a reverse map, which must
// be implemented by the BBHash type.
type reverseMap interface {
	// Key returns the key for the given index.
	Key(index uint64) uint64
}
