package fast

import (
	"crypto/rand"
	"encoding/binary"
)

const m uint64 = 0x880355f21e6d1965

// Hash returns the hash of the current level and key.
func Hash(level, key uint64) uint64 {
	return KeyHash(LevelHash(level), key)
}

// LevelHash returns the hash of the given level.
func LevelHash(level uint64) uint64 {
	return mix(level) * m
}

// KeyHash returns the hash of a key given a level hash.
func KeyHash(levelHash, key uint64) uint64 {
	var h uint64 = levelHash
	h ^= mix(key)
	h *= m
	h = mix(h)
	return h
}

// mix is a compression function for fast hashing.
func mix(h uint64) uint64 {
	h ^= h >> 23
	h *= 0x2127599bf4325c37
	h ^= h >> 47
	return h
}

// rand64 returns a 64-bit cryptographic random number.
func rand64() (rnd uint64) {
	binary.Read(rand.Reader, binary.LittleEndian, &rnd)
	return
}
