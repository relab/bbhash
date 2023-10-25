package bbhash

type ReverseMap [][]byte

// ReverseMap returns a reverse map from the proof's MPHF index to
// the chunk's content address (or database key).
func (bb *BBHash) ReverseMap(dbKeys [][]byte, mphfKeys []uint64) ReverseMap {
	// len+1 since the first index represents not-found.
	reverseMap := make(ReverseMap, len(mphfKeys)+1)
	for i, cp := range mphfKeys {
		index := bb.Find(cp)
		reverseMap[index] = dbKeys[i]
	}
	return reverseMap
}

// Lookup returns the chunk's content address (or database key) for the given MPHF index.
func (rm ReverseMap) Lookup(index uint64) []byte {
	if index >= uint64(len(rm)) {
		return nil
	}
	return rm[index]
}
