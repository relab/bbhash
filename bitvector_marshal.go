package bbhash

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// marshaledLength returns the number of bytes needed to marshal the bit vector.
func (b *bitVector) marshaledLength() int {
	const uint64bytes = 8
	return uint64bytes * (1 + len(b.v))
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *bitVector) MarshalBinary() ([]byte, error) {
	out := make([]byte, b.marshaledLength())
	// Make a copy of out, since we will be modifying buf's slice indices
	buf := out

	// Write the number of uint64 words needed for this bit vector to the out buffer
	binary.LittleEndian.PutUint64(buf[:8], b.words())
	buf = buf[8:]

	// Write the bit vector entries to the out buffer
	for _, v := range b.v {
		binary.LittleEndian.PutUint64(buf[:8], v)
		buf = buf[8:]
	}

	return out, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *bitVector) UnmarshalBinary(data []byte) error {
	// Make a copy of data, since we will be modifying buf's slice indices
	buf := data
	if len(buf) < 8 {
		return errors.New("bitVector.UnmarshalBinary: no data")
	}

	// Read the number of words in the bit vector
	words := binary.LittleEndian.Uint64(buf[:8])
	if words == 0 || words > (1<<32) {
		return fmt.Errorf("bitVector.UnmarshalBinary: invalid length %d", words)
	}

	b.v = make([]uint64, words)
	buf = buf[8:]

	// Read the bit vector entries
	for i := uint64(0); i < words; i++ {
		if len(buf) < 8 {
			return errors.New("bitVector.UnmarshalBinary: not enough data to read bit vector entry")
		}
		b.v[i] = binary.LittleEndian.Uint64(buf[:8])
		buf = buf[8:]
	}

	return nil
}
