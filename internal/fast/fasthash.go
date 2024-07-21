package fast

import (
	"unsafe"
)

// The code below is an improved version of the code at <https://github.com/opencoff/go-fasthash>

// hash64 returns the hash of the given buffer.
func hash64(seed uint64, buf []byte) uint64 {
	h := seed ^ (uint64(len(buf)) * m)

	if n := len(buf) / 8; n > 0 {
		// Convert []byte to []uint64 using unsafe.Slice
		data := unsafe.Slice((*uint64)(unsafe.Pointer(&buf[0])), n)
		// Mix 8 bytes at a time
		for _, v := range data {
			h ^= mix(v)
			h *= m
		}
		buf = buf[n*8:]
	}

	var v uint64
	for i, b := range buf {
		v |= uint64(b) << (8 * uint(i))
	}
	if len(buf) > 0 {
		h ^= mix(v)
		h *= m
	}
	return mix(h)
}
