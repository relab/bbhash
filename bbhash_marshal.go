package bbhash

import (
	"encoding/binary"
	"errors"
	"fmt"
)

func (bb BBHash) marshaledLength() int {
	words := 1 // one word for header
	for _, bv := range bb.bits {
		words += int(bv.words())
	}
	return uint64bytes * words
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
func (bb BBHash) AppendBinary(buf []byte) ([]byte, error) {
	// number of bit vectors (= number of levels)
	numBitVectors := uint64(len(bb.bits))
	if numBitVectors == 0 {
		return nil, errors.New("BBHash.AppendBinary: no data")
	}
	// append header: the number of bit vectors
	buf = binary.LittleEndian.AppendUint64(buf, numBitVectors)

	var err error
	// append the bit vector for each level
	for _, bv := range bb.bits {
		buf, err = bv.AppendBinary(buf)
		if err != nil {
			return nil, err
		}
	}

	// We don't append the rank vector, since we can re-compute it
	// when we unmarshal the bit vectors.
	// Similarly, the reverse map it is not meant to be serialized.

	return buf, nil
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (bb BBHash) MarshalBinary() ([]byte, error) {
	return bb.AppendBinary(make([]byte, 0, bb.marshaledLength()))
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (bb *BBHash) UnmarshalBinary(data []byte) error {
	// Make a copy of data, since we will be modifying buf's slice indices
	buf := data
	if len(buf) < uint64bytes {
		return errors.New("BBHash.UnmarshalBinary: no data")
	}

	// Read header: the number of bit vectors
	numBitVectors := binary.LittleEndian.Uint64(buf[:uint64bytes])
	if numBitVectors == 0 || numBitVectors > maxLevel {
		return fmt.Errorf("BBHash.UnmarshalBinary: invalid number of bit vectors %d (max %d)", numBitVectors, maxLevel)
	}

	*bb = BBHash{} // modify bb in place
	bb.bits = make([]bitVector, numBitVectors)
	buf = buf[uint64bytes:] // move past header

	// Read bit vectors for each level
	for i := range numBitVectors {
		bv := bitVector{}
		if err := bv.UnmarshalBinary(buf); err != nil {
			return err
		}
		bb.bits[i] = bv
		bvLen := bv.marshaledLength()
		if len(buf) < bvLen {
			return errors.New("BBHash.UnmarshalBinary: insufficient data for remaining bit vectors")
		}
		buf = buf[bvLen:] // move past the current bit vector
	}

	bb.computeLevelRanks()
	return nil
}
