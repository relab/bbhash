package bbhash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (bb BBHash) MarshalBinary() ([]byte, error) {
	// Header: 1 x 64-bit words:
	//   n - number of bit vectors (= number of levels)
	//
	// Body:
	//   <n> bit vectors laid out consecutively

	numBitVectors := uint64(len(bb.bits))
	if numBitVectors == 0 {
		return nil, errors.New("BBHash.MarshalBinary: invalid length")
	}

	// Write header
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, numBitVectors); err != nil {
		return nil, err
	}
	// Write bit vectors for each level
	for _, bv := range bb.bits {
		b, err := bv.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(b); err != nil {
			return nil, err
		}
	}

	// We don't store the rank vector, since we can re-compute it
	// when we unmarshal the bit vectors.

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (bb *BBHash) UnmarshalBinary(data []byte) error {
	// Make a copy of data, since we will be modifying buf's slice indices
	buf := data
	if len(buf) < 8 {
		return errors.New("BBHash.UnmarshalBinary: no data")
	}

	// Read header
	numBitVectors := binary.LittleEndian.Uint64(buf[:8])
	if numBitVectors == 0 || numBitVectors > maxLevel {
		return fmt.Errorf("BBHash.UnmarshalBinary: invalid number of bit vectors %d (max %d)", numBitVectors, maxLevel)
	}
	*bb = BBHash{} // Not strictly necessary, but seems to be recommended practice
	bb.bits = make([]*bitVector, numBitVectors)
	buf = buf[8:] // Move past header

	// Read bit vectors for each level
	for i := uint64(0); i < numBitVectors; i++ {
		bv := &bitVector{}
		if err := bv.UnmarshalBinary(buf); err != nil {
			return err
		}
		bb.bits[i] = bv
		bvLen := bv.marshaledLength()
		if len(buf) < bvLen {
			return errors.New("BBHash.UnmarshalBinary: insufficient data for remaining bit vectors")
		}
		buf = buf[bvLen:] // Move past the current bit vector
	}

	bb.computeLevelRanks()
	return nil
}
