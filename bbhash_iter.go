package bbhash

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"iter"

	"github.com/relab/bbhash/internal/fast"
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

var SHA256HashFunc = func(buf []byte) uint64 {
	h := sha256.New()
	h.Write(buf)
	return binary.LittleEndian.Uint64(h.Sum(nil))
}

var FastHashFunc = func(buf []byte) uint64 {
	return fast.Hash64(123, buf)
}
