package bbhash

func NewParallel(gamma float64, salt uint64, keys []uint64) (*BBHash, error) {
	// TODO: this is just a placeholder for now
	return NewSequential(gamma, salt, keys)
}
