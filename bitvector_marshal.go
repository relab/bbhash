package bbhash

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const uint64bytes = 8

// marshaledLength returns the number of bytes needed to marshal the bit vector.
func (b bitVector) marshaledLength() int {
	return uint64bytes * (1 + len(b))
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
func (b bitVector) AppendBinary(buf []byte) ([]byte, error) {
	// append the number of words needed for this bit vector to the buffer
	buf = binary.LittleEndian.AppendUint64(buf, b.words())
	// append the bit vector entries to the buffer
	for _, v := range b {
		buf = binary.LittleEndian.AppendUint64(buf, v)
	}
	return buf, nil
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (b bitVector) MarshalBinary() ([]byte, error) {
	return b.AppendBinary(make([]byte, 0, b.marshaledLength()))
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (b *bitVector) UnmarshalBinary(data []byte) error {
	// Make a copy of data, since we will be modifying buf's slice indices
	buf := data
	if len(buf) < uint64bytes {
		return errors.New("bitVector.UnmarshalBinary: no data")
	}

	// Read the number of words in the bit vector
	words := binary.LittleEndian.Uint64(buf[:uint64bytes])
	if words == 0 || words > (1<<32) {
		return fmt.Errorf("bitVector.UnmarshalBinary: invalid bit vector length %d (max %d)", words, 1<<32)
	}
	buf = buf[uint64bytes:] // move past header

	*b = make(bitVector, words) // modify b in place

	// Read the bit vector entries
	for i := range words {
		if len(buf) < uint64bytes {
			return errors.New("bitVector.UnmarshalBinary: insufficient data for bit vector entry")
		}
		(*b)[i] = binary.LittleEndian.Uint64(buf[:uint64bytes])
		buf = buf[uint64bytes:]
	}

	return nil
}
