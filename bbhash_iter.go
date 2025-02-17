package bbhash

import (
	"io"
	"iter"
)

// Find the chunks from slow memory
// Chunks returns smaller chunks of data given a reader with some data already being read and with the buffer size.
func ReadChunks(readerInfo io.Reader, bufSz int) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		buffer := make([]byte, bufSz)
		for {
			// Create buffer and read the input into it for a certain range of bytes
			n, err := readerInfo.Read(buffer)
			if err != nil {
				return
			}
			if !yield(buffer[:n]) {
				return
			}
		}
	}
}

// Keys returns the hashes of the chunks using the provided hash function
func Keys(hashFunc func([]byte) uint64, chunks iter.Seq[[]byte]) []uint64 {
	var keys []uint64
	for c := range chunks {
		keys = append(keys, hashFunc(c))
	}

	return keys
}
