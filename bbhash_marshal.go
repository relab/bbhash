package bbhash

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// marshalLength returns the number of bytes needed to marshal the BBHash.
func (bb BBHash) marshaledLength() int {
	bbLen := 1 // one byte for header: max 255 levels
	for _, bv := range bb.bits {
		bbLen += bv.marshaledLength()
	}
	return bbLen
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
func (bb BBHash) AppendBinary(buf []byte) (_ []byte, err error) {
	numBitVectors := uint8(len(bb.bits))
	if numBitVectors == 0 {
		return nil, errors.New("BBHash.AppendBinary: no data")
	}
	// append header: the number of bit vectors (levels)
	buf = append(buf, numBitVectors)

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
	if len(buf) < 1 {
		return errors.New("BBHash.UnmarshalBinary: no data")
	}

	// Read header: the number of bit vectors
	numBitVectors := uint8(buf[0])
	if numBitVectors == 0 || numBitVectors > maxLevel {
		return fmt.Errorf("BBHash.UnmarshalBinary: invalid number of bit vectors %d (max %d)", numBitVectors, maxLevel)
	}
	buf = buf[1:] // move past header

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

// marshalLength returns the number of bytes needed to marshal the BBHash2.
func (b2 BBHash2) marshaledLength() int {
	b2Len := 1 // one byte for header: max 255 partitions
	// length of each partition
	for _, bb := range b2.partitions {
		b2Len += bb.marshaledLength()
	}
	// length of the offset vector (excluding the first offset which is always 0)
	b2Len += uint32bytes * (len(b2.offsets) - 1)
	return b2Len
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
func (b2 BBHash2) AppendBinary(buf []byte) (_ []byte, err error) {
	numPartitions := uint8(len(b2.partitions))
	if numPartitions == 0 {
		return nil, errors.New("BBHash2.AppendBinary: no data")
	}
	// append header: the number of partitions
	buf = append(buf, numPartitions)

	// append the BBHash for each partition
	for _, bb := range b2.partitions {
		buf, err = bb.AppendBinary(buf)
		if err != nil {
			return nil, err
		}
	}
	// append the offset vector (excluding the first offset which is always 0)
	for i := 1; i < len(b2.offsets); i++ {
		buf = binary.LittleEndian.AppendUint32(buf, b2.offsets[i])
	}

	return buf, nil
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (b2 BBHash2) MarshalBinary() ([]byte, error) {
	return b2.AppendBinary(make([]byte, 0, b2.marshaledLength()))
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (b2 *BBHash2) UnmarshalBinary(data []byte) error {
	// Make a copy of data, since we will be modifying buf's slice indices
	buf := data
	if len(buf) < 1 {
		return errors.New("BBHash2.UnmarshalBinary: no data")
	}

	// Read header: the number of partitions
	numPartitions := uint8(buf[0])
	if numPartitions == 0 || numPartitions > maxPartitions {
		return fmt.Errorf("BBHash2.UnmarshalBinary: invalid number of partitions %d (max %d)", numPartitions, maxPartitions)
	}
	buf = buf[1:] // move past header

	*b2 = BBHash2{} // modify b2 in place
	b2.partitions = make([]BBHash, numPartitions)

	// Read BBHash for each partition
	for i := range numPartitions {
		bb := BBHash{}
		if err := bb.UnmarshalBinary(buf); err != nil {
			return err
		}
		b2.partitions[i] = bb
		bbLen := bb.marshaledLength()
		if len(buf) < bbLen {
			return errors.New("BBHash2.UnmarshalBinary: insufficient data for remaining partitions")
		}
		buf = buf[bbLen:] // move past the current partition
	}

	// we skip the first offset since it is always 0, hence numPartitions-1
	if len(buf) < int(uint32bytes*(numPartitions-1)) {
		return errors.New("BBHash2.UnmarshalBinary: insufficient data for offset vector")
	}

	// Read offset vector
	b2.offsets = make([]uint32, numPartitions)
	b2.offsets[0] = 0 // first offset is always 0
	for i := uint8(1); i < numPartitions; i++ {
		b2.offsets[i] = binary.LittleEndian.Uint32(buf[:uint32bytes])
		buf = buf[uint32bytes:] // move past the current offset
	}

	return nil
}
