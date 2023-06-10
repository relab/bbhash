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
				lvlSz = append(lvlSz, bv.size())
				lvlEntries = append(lvlEntries, bv.onesCount())
			} else {
				lvlSz[lvl] += bv.size()
				lvlEntries[lvl] += bv.onesCount()
			}
		}
	}
	b.WriteString(fmt.Sprintf("BBHash2(entries=%d, levels=%d, bits=%d, size=%s, bits per key=%3.1f)\n",
		bb.entries(), len(lvlSz), bb.numBits(), bb.space(), bb.BitsPerKey()))
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
	return float64(bb.numBits()) / float64(bb.entries())
}

func (bb BBHash2) space() string {
	return readableSize(bb.numBits() / 8)
}

func (bb BBHash2) entries() uint64 {
	var sz uint64
	for _, bx := range bb.partitions {
		sz += bx.entries()
	}
	return sz
}

func (bb BBHash2) numBits() (sz uint64) {
	const sizeOfInt = 64
	for _, bx := range bb.partitions {
		sz += bx.numBits()
		sz += sizeOfInt // to account for the offset
	}
	return sz
}
