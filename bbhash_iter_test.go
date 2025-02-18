package bbhash_test

import (
	"crypto/sha256"
	"encoding/binary"
	"iter"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/relab/bbhash"
	"github.com/relab/bbhash/internal/fast"
)

// String taken from https://www.lipsum.com/
const input string = "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."

func TestChunks(t *testing.T) {
	// Read the string and get the resulting bytes from the Chunks() method
	r := strings.NewReader(input)
	bufSz := 128
	wantChunks := slices.Collect(slices.Chunk([]byte(input), bufSz))

	i := 0
	for got := range bbhash.ReadChunks(r, bufSz) {
		if diff := cmp.Diff(got, wantChunks[i]); diff != "" {
			t.Errorf("Chunks() (-got +want)\n%s", diff)
		}
		i++
	}
}

var sha256hashFunc = func(buf []byte) uint64 {
	h := sha256.New()
	h.Write(buf)
	return binary.LittleEndian.Uint64(h.Sum(nil))
}

var fastHashFunc = func(buf []byte) uint64 {
	return fast.Hash64(uint64(123), buf)
}

func CollectFunc[I, O any](seq iter.Seq[I], f func(I) O) (o []O) {
	for v := range seq {
		o = append(o, f(v))
	}
	return
}

func TestHashKeysFromChunks(t *testing.T) {
	tests := []struct {
		name      string
		hashFunc  func([]byte) uint64
		in        string
		chunkSize int
	}{
		{name: "FashHash", hashFunc: fastHashFunc, in: input[:5], chunkSize: 4},
		{name: "FashHash", hashFunc: fastHashFunc, in: input[:5], chunkSize: 8},
		{name: "SHA256", hashFunc: sha256hashFunc, in: input[:5], chunkSize: 4},
		{name: "SHA256", hashFunc: sha256hashFunc, in: input[:5], chunkSize: 8},
		{name: "LongFast", hashFunc: fastHashFunc, in: input, chunkSize: 128},
		{name: "LongSHA", hashFunc: sha256hashFunc, in: input, chunkSize: 128},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Prepare the expected hashed keys
			wantHashedKeys := CollectFunc(slices.Chunk([]byte(test.in), test.chunkSize), func(v []byte) uint64 {
				return test.hashFunc(v)
			})

			r := strings.NewReader(test.in)
			chunks := bbhash.ReadChunks(r, test.chunkSize)
			gotHashedKeys := bbhash.Keys(test.hashFunc, chunks)

			if diff := cmp.Diff(gotHashedKeys, wantHashedKeys); diff != "" {
				t.Errorf("Keys(): (-got +want) \n%s", diff)
			}
		})
	}
}
