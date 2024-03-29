package bbhash

import (
	"fmt"
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
