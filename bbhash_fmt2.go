package bbhash

import (
	"fmt"
	"go/format"
	"strings"
)

func (bb BBHash2) String() string {
	var b strings.Builder
	lvlSz := make([]uint64, 0)
	lvlEntries := make([]uint64, 0)
	for _, bx := range bb.partitions {
		for lvl, bv := range bx.bits {
			if lvl >= len(lvlSz) {
				// first partition at lvl; start new counters for size and entries
				lvlSz = append(lvlSz, bv.size())
				lvlEntries = append(lvlEntries, bv.onesCount())
			} else {
				lvlSz[lvl] += bv.size()
				lvlEntries[lvl] += bv.onesCount()
			}
		}
	}
	b.WriteString(fmt.Sprintf("BBHash2(entries=%d, levels=%d, bits per key=%3.1f, wire bits=%d, size=%s)\n",
		bb.entries(), len(lvlSz), bb.BitsPerKey(), bb.wireBits(), bb.space()))
	for lvl := 0; lvl < len(lvlSz); lvl++ {
		sz := int(lvlSz[lvl])
		entries := lvlEntries[lvl]
		b.WriteString(fmt.Sprintf("  %d: %d / %d bits (%s)\n", lvl, entries, sz, readableSize(sz/8)))
	}
	return b.String()
}

// MaxMinLevels returns the maximum and minimum number of levels across all partitions.
func (bb BBHash2) MaxMinLevels() (max, min int) {
	max = 0
	min = 999
	for _, bx := range bb.partitions {
		if max < bx.Levels() {
			max = bx.Levels()
		}
		if min > bx.Levels() {
			min = bx.Levels()
		}
	}
	return max, min
}

// BitsPerKey returns the number of bits per key in the minimal perfect hash.
func (bb BBHash2) BitsPerKey() float64 {
	return float64(bb.wireBits()) / float64(bb.entries())
}

// LevelVectors returns a slice representation of BBHash2's per-partition, per-level bit vectors.
func (bb BBHash2) LevelVectors() [][][]uint64 {
	var vectors [][][]uint64
	for _, bx := range bb.partitions {
		vectors = append(vectors, bx.LevelVectors())
	}
	return vectors
}

// BitVectors returns a Go slice for BBHash2's per-partition, per-level bit vectors.
// This is intended for testing and debugging; no guarantees are made about the format.
func (bb BBHash2) BitVectors(varName string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("var %s = [][][]uint64{\n", varName))
	for partition, bx := range bb.partitions {
		b.WriteString(fmt.Sprintf("// Partition %d:\n{\n", partition))
		for lvl, bv := range bx.bits {
			b.WriteString(fmt.Sprintf("// Level %d:\n{\n", lvl))
			for _, v := range bv {
				b.WriteString(fmt.Sprintf("%#016x,\n", v))
			}
			b.WriteString("},\n")
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

// entries returns the number of entries in the minimal perfect hash.
func (bb BBHash2) entries() (sz uint64) {
	for _, bx := range bb.partitions {
		sz += bx.entries()
	}
	return sz
}

// wireBits returns the number of on-the-wire bits used to represent the minimal perfect hash.
func (bb BBHash2) wireBits() uint64 {
	return uint64(bb.marshaledLength()) * 8
}

// space returns a human-readable string representing the size of the minimal perfect hash.
func (bb BBHash2) space() string {
	return readableSize(bb.marshaledLength())
}
