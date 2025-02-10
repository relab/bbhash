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
	b.WriteString(fmt.Sprintf("BBHash2(entries=%d, levels=%d, mem bits=%d, wire bits=%d, size=%s, bits per key=%3.1f)\n",
		bb.entries(), len(lvlSz), bb.numBits(), bb.wireBits(), bb.space(), bb.BitsPerKey()))
	for lvl := 0; lvl < len(lvlSz); lvl++ {
		sz := lvlSz[lvl]
		entries := lvlEntries[lvl]
		b.WriteString(fmt.Sprintf("  %d: %d / %d bits (%s)\n", lvl, entries, sz, readableSize(sz/8)))
	}
	return b.String()
}

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

func (bb BBHash2) BitsPerKey() float64 {
	return float64(bb.wireBits()) / float64(bb.entries())
}

func (bb BBHash2) space() string {
	return readableSize(bb.wireBits() / 8)
}

func (bb BBHash2) entries() uint64 {
	var sz uint64
	for _, bx := range bb.partitions {
		sz += bx.entries()
	}
	return sz
}

// numBits returns the number of in-memory bits used to represent the minimal perfect hash.
func (bb BBHash2) numBits() (sz uint64) {
	const sizeOfInt = 64
	for _, bx := range bb.partitions {
		sz += bx.numBits()
		sz += sizeOfInt // to account for the offset
	}
	return sz
}

// wireBits returns the number of on-the-wire bits used to represent the minimal perfect hash on the wire.
func (bb BBHash2) wireBits() (sz uint64) {
	if len(bb.partitions) == 1 {
		// no need to account for the offset since there is only one partition
		return bb.partitions[0].wireBits()
	}
	const sizeOfInt = 64
	for _, bx := range bb.partitions {
		sz += bx.wireBits()
		sz += sizeOfInt // to account for the offset
	}
	return sz
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

// LevelVectors returns a slice representation of BBHash's per-partition, per-level bit vectors.
func (bb BBHash2) LevelVectors() [][][]uint64 {
	var vectors [][][]uint64
	for _, bx := range bb.partitions {
		vectors = append(vectors, bx.LevelVectors())
	}
	return vectors
}
