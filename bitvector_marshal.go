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
		return fmt.Errorf("bitVector.UnmarshalBinary: invalid length %d", words)
	}

	*b = make(bitVector, words) // modify b in place
	buf = buf[uint64bytes:]

	// Read the bit vector entries
	for i := uint64(0); i < words; i++ {
		if len(buf) < uint64bytes {
			return errors.New("bitVector.UnmarshalBinary: not enough data to read bit vector entry")
		}
		(*b)[i] = binary.LittleEndian.Uint64(buf[:uint64bytes])
		buf = buf[uint64bytes:]
	}

	return nil
}
