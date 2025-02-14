package bbhash

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// marshalLength returns the number of bytes needed to marshal the BBHash.
func (bb BBHash) marshaledLength() int {
	bbLen := uint64bytes // one word for header
	for _, bv := range bb.bits {
		bbLen += bv.marshaledLength()
	}
	return bbLen
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
func (bb BBHash) AppendBinary(buf []byte) (_ []byte, err error) {
	numBitVectors := uint64(len(bb.bits))
	if numBitVectors == 0 {
		return nil, errors.New("BBHash.AppendBinary: no data")
	}
	// append header: the number of bit vectors (levels)
	buf = binary.LittleEndian.AppendUint64(buf, numBitVectors)

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
	buf = buf[uint64bytes:] // move past header

	*bb = BBHash{} // modify bb in place
	bb.bits = make([]bitVector, numBitVectors)

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
