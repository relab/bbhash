package bbhash

import (
	"fmt"
	"go/format"
	"strings"
)

// String returns a string representation of BBHash with overall and per-level statistics.
func (bb BBHash) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("BBHash(gamma=%3.1f, entries=%d, levels=%d, bits per key=%3.1f, wire bits=%d, size=%s, false positive rate=%.2f)\n",
		bb.gamma(), bb.entries(), bb.Levels(), bb.BitsPerKey(), bb.wireBits(), bb.space(), bb.falsePositiveRate()))
	for i, bv := range bb.bits {
		sz := readableSize(int(bv.words()) * 8)
		entries := bv.onesCount()
		b.WriteString(fmt.Sprintf("  %d: %d / %d bits (%s)\n", i, entries, bv.size(), sz))
	}
	return b.String()
}

// Levels returns the number of Levels in the minimal perfect hash.
func (bb BBHash) Levels() int {
	return len(bb.bits)
}

// BitsPerKey returns the number of bits per key in the minimal perfect hash.
func (bb BBHash) BitsPerKey() float64 {
	return float64(bb.wireBits()) / float64(bb.entries())
}

// LevelVectors returns a slice representation of the BBHash's per-level bit vectors.
func (bb BBHash) LevelVectors() [][]uint64 {
	m := make([][]uint64, 0, len(bb.bits))
	for _, bv := range bb.bits {
		m = append(m, bv)
	}
	return m
}

// BitVectors returns a Go slice for BBHash's per-level bit vectors.
// This is intended for testing and debugging; no guarantees are made about the format.
func (bb BBHash) BitVectors(varName string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("var %s = [][]uint64{\n", varName))
	for lvl, bv := range bb.bits {
		b.WriteString(fmt.Sprintf("// Level %d:\n{\n", lvl))
		for _, v := range bv {
			b.WriteString(fmt.Sprintf("%#016x,\n", v))
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	s, err := format.Source([]byte(b.String()))
	if err != nil {
		panic(err)
	}
	return string(s)
}

const (
	_B = 1 << (iota * 10)
	_kB
	_MB
	_GB
	_TB
	_PB
)

// readableSize returns a human readable representation of the size in bytes.
func readableSize(sizeInBytes int) string {
	sz := float64(sizeInBytes)
	switch {
	case sizeInBytes >= _PB:
		return fmt.Sprintf("%3.1f PB", sz/_PB)
	case sizeInBytes >= _TB:
		return fmt.Sprintf("%3.1f TB", sz/_TB)
	case sizeInBytes >= _GB:
		return fmt.Sprintf("%3.1f GB", sz/_GB)
	case sizeInBytes >= _MB:
		return fmt.Sprintf("%3.1f MB", sz/_MB)
	case sizeInBytes >= _kB:
		return fmt.Sprintf("%3.1f KB", sz/_kB)
	}
	return fmt.Sprintf("%d B", sizeInBytes)
}

// gamma returns an estimate of the gamma parameter used to construct the minimal perfect hash.
// It is an estimate because the size of the level 0 bit vector is not necessarily a multiple of 64.
func (bb BBHash) gamma() float64 {
	lvl0Size := bb.bits[0].size()
	return float64(lvl0Size) / float64(bb.entries())
}

// entries returns the number of entries in the minimal perfect hash.
func (bb BBHash) entries() (sz uint64) {
	for _, bv := range bb.bits {
		sz += bv.onesCount()
	}
	return sz
}

// wireBits returns the number of on-the-wire bits used to represent the minimal perfect hash.
func (bb BBHash) wireBits() uint64 {
	return uint64(bb.marshaledLength()) * 8
}

// space returns a human-readable string representing the size of the minimal perfect hash.
func (bb BBHash) space() string {
	return readableSize(bb.marshaledLength())
}

// falsePositiveRate returns the false positive rate of the minimal perfect hash.
// Note: This may not be accurate if the actual keys overlap with the test keys [0,2N];
// that is, if many of the actual keys are in the range [0,2N], then it will be inaccurate.
func (bb BBHash) falsePositiveRate() float64 {
	var cnt int
	numTestKeys := bb.entries() * 2
	for key := uint64(0); key < numTestKeys; key++ {
		hashIndex := bb.Find(key)
		if hashIndex != 0 {
			cnt++
		}
	}
	return float64(cnt) / float64(numTestKeys)
}
