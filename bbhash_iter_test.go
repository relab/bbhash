package bbhash_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/relab/bbhash"
	"github.com/relab/bbhash/internal/fast"
)

func TestChunks(t *testing.T) {
	input := "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."
	// Read the string and get the resulting bytes from the Chunks() method
	r := strings.NewReader(input)
	bufSz := 128
	wantChunks := slices.Collect(slices.Chunk([]byte(input), bufSz))

	i := 0
	for got := range bbhash.ReadChunks(r, bufSz) {
		// Remove empty slots to make the "got" and "want" chunks to be the same,
		//  in case the whole chunk of bytes is not used.
		got = bytes.Trim(got, "\x00")
		if diff := cmp.Diff(got, wantChunks[i]); diff != "" {
			t.Errorf("Chunks() (-got +want)\n%s", diff)
		}
		i++
	}
}

func TestHashKeysFromChunksSingleWord(t *testing.T) {
	word := "Hello"
	hashFunc := 0 // 0 = Fasthash, 1 = sha256
	r := strings.NewReader(word)
	chunks := bbhash.ReadChunks(r, 8)

	gotHashedKeys := bbhash.Keys(hashFunc, chunks)

	// As we are dealing with an append on wordByte, som bytes are removed...
	// ...from the initial creation such that wordByte and chunks have the same length
	wordByte := make([]byte, 3)
	wordByte = append([]byte(word), wordByte...)
	wantHashedKeys := []uint64{fast.Hash64(1, wordByte)}

	for i := range wantHashedKeys {
		if diff := cmp.Diff(gotHashedKeys[i], wantHashedKeys[i]); diff != "" {
			t.Errorf("Keys(): (-got +want) \n%s", diff)
		}
	}
}

func TestHashKeysFromChunksSingleWordSHA256(t *testing.T) {
	word := "Hello"
	hashFunc := 1 // 0 = Fasthash, 1 = sha256
	r := strings.NewReader(word)
	chunks := bbhash.ReadChunks(r, 8)

	gotHashedKeys := bbhash.Keys(hashFunc, chunks)

	// As we are dealing with an append on wordByte, som bytes are removed...
	// ...from the initial creation such that wordByte and chunks have the same length
	wordByte := make([]byte, 3)
	wordByte = append([]byte(word), wordByte...)
	h := sha256.New()
	h.Write(wordByte)
	s := binary.LittleEndian.Uint64(h.Sum(nil))
	wantHashedKeys := []uint64{s}

	for i := range wantHashedKeys {
		if diff := cmp.Diff(gotHashedKeys[i], wantHashedKeys[i]); diff != "" {
			t.Errorf("Keys(): (-got +want) \n%s", diff)
		}
	}
}

func TestHashKeysFromChunksLongerWords(t *testing.T) {
	// String taken from https://www.lipsum.com/
	input := "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."
	hashFunc := 0 // 0 = Fasthash, 1 = sha256
	bufSz := 128

	r := strings.NewReader(input)
	chunks := bbhash.ReadChunks(r, bufSz)
	gotKeyHashes := bbhash.Keys(hashFunc, chunks)

	wantChunks := slices.Collect(slices.Chunk([]byte(input), bufSz))
	wantKeyHashes := []uint64{}

	// Last hash element might containt zeroes, meaning we have to append some zeroes  to the original input
	// so that both will produce the same hash
	if len(wantChunks[len(wantChunks)-1]) != bufSz {
		for i := len(wantChunks[len(wantChunks)-1]); i < bufSz; i++ {
			wantChunks[len(wantChunks)-1] = append(wantChunks[len(wantChunks)-1][:], 0)
		}
	}

	// Hash the keys using fast.Hash64() using 1 as the seed, which was chosen arbitrarily, to compare to the keys from Keys()
	for i := range wantChunks {
		wordHash := fast.Hash64(1, []byte(wantChunks[i]))
		wantKeyHashes = append(wantKeyHashes, wordHash)
	}

	for i := range wantKeyHashes {
		if diff := cmp.Diff(gotKeyHashes[i], wantKeyHashes[i]); diff != "" {
			t.Errorf("Keys() (-got +want) \n%s", diff)
		}

	}
}

func TestHashKeysFromChunksLongerWordsSHA256(t *testing.T) {
	// String taken from https://www.lipsum.com/
	input := "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."
	hashFunc := 1 // 0 = Fasthash, 1 = sha256
	bufSz := 128

	r := strings.NewReader(input)
	chunks := bbhash.ReadChunks(r, bufSz)
	gotKeyHashes := bbhash.Keys(hashFunc, chunks)

	wantChunks := slices.Collect(slices.Chunk([]byte(input), bufSz))
	wantKeyHashes := []uint64{}

	// Last hash element might containt zeroes, meaning we have to append some zeroes  to the original input
	// so that both will produce the same hash
	if len(wantChunks[len(wantChunks)-1]) != bufSz {
		for i := len(wantChunks[len(wantChunks)-1]); i < bufSz; i++ {
			wantChunks[len(wantChunks)-1] = append(wantChunks[len(wantChunks)-1][:], 0)
		}
	}

	// Hash the keys using SHA256 and add them to the slice to compare to hte keys from Keys()
	for i := range wantChunks {
		h := sha256.New()
		h.Write(wantChunks[i])
		s := binary.LittleEndian.Uint64(h.Sum(nil))
		wantKeyHashes = append(wantKeyHashes, s)
	}

	for i := range wantKeyHashes {
		if diff := cmp.Diff(gotKeyHashes[i], wantKeyHashes[i]); diff != "" {
			t.Errorf("Keys() (-got +want) \n%s", diff)
		}

	}
}
