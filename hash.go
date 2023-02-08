package bbhash

import (
	"crypto/rand"
	"encoding/binary"
)

const m uint64 = 0x880355f21e6d1965

// hash returns the hash of a salt, level, and key.
func hash(saltHash, lvl, key uint64) uint64 {
	return keyHash(levelHash(saltHash, lvl), key)
}

// saltHash returns the hash of a salt.
func saltHash(salt uint64) uint64 {
	var h uint64 = m
	h ^= mix(salt)
	h *= m
	return h
}

// levelHash returns the hash of a level given a salt hash.
func levelHash(saltHash, lvl uint64) uint64 {
	var h uint64 = saltHash
	h ^= mix(lvl)
	h *= m
	return h
}

// keyHash returns the hash of a key given a level hash.
func keyHash(levelHash, key uint64) uint64 {
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
