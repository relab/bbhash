package bbhash

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"iter"

	"github.com/relab/bbhash/internal/fast"
)

type BBHashSeq struct {
	bits       []*bitVector
	ranks      []uint64
	reverseMap []uint64
	hashFunc   int // Use int instead of bool incase there can be more hash functions tested than fasthash64 and SHA256. 0 = hash64, 1 = SHA256.
	MaxLevel   int
}

// Find the chunks from slow memory
// Currently only takes the whole sequence of input and crashes if it's not able to fit everything in the buffer.
func ReadChunks(readerInfo io.Reader, bufSz int) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		i := 0
		for {
			// Create buffer and read the input into it for a certain range of bytes
			buffer := make([]byte, bufSz)
			_, err := readerInfo.Read(buffer)
			if err != nil {
				return
			}
			i++
			if !yield(buffer) {
				return
			}
		}
	}
}

// Int = hashfunc, iter.Seq chunks.
// Recieves chunks from chunks() and converts them to keys and returns them as []uint64
func Keys(hashFunc int, chunks iter.Seq[[]byte]) []uint64 {
	var keys []uint64
	if hashFunc == 0 {
		for c := range chunks {
			keys = append(keys, fast.Hash64(1, c)) //1 is just a magic number
		}
	} else {
		for c := range chunks {
			h := sha256.New()
			h.Write(c)
			s := binary.LittleEndian.Uint64(h.Sum(nil))
			keys = append(keys, s)
		}
	}

	return keys
}
