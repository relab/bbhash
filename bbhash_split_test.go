package bbhash

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestSplit(t *testing.T) {
	parts := 16
	tests := []struct {
		size          int
		expectedParts int
	}{
		{size: 10, expectedParts: 1},
		{size: 100, expectedParts: 2},
		{size: 200, expectedParts: 4},
		{size: 300, expectedParts: 5},
		{size: 400, expectedParts: 7},
		{size: 500, expectedParts: 8},
		{size: 512, expectedParts: 8},
		{size: 513, expectedParts: 9},
		{size: 1000, expectedParts: 16},
		{size: 10_000, expectedParts: 16},
		{size: 100_000, expectedParts: 16},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("size=%d", tt.size), func(t *testing.T) {
			keys := generateKeys(tt.size, 99)
			k := splitP(keys, parts)
			if len(k) != tt.expectedParts {
				t.Errorf("len(keys) = %d ==> len(k) = %d, want %d", len(keys), len(k), tt.expectedParts)
			}
		})
	}
}

func TestSplitX(t *testing.T) {
	t.Skip("Sensitive to changes in the w const in splitX()")
	tests := []struct {
		size          int
		expectedParts int
	}{
		{size: 10, expectedParts: 1},
		{size: 100, expectedParts: 2},
		{size: 200, expectedParts: 4},
		{size: 300, expectedParts: 5},
		{size: 400, expectedParts: 7},
		{size: 500, expectedParts: 8},
		{size: 512, expectedParts: 8},
		{size: 513, expectedParts: 9},
		{size: 1000, expectedParts: 16},
		{size: 10_000, expectedParts: 157},
		{size: 100_000, expectedParts: 1563},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("size=%d", tt.size), func(t *testing.T) {
			keys := generateKeys(tt.size, 99)
			k := splitX(keys)
			if len(k) != tt.expectedParts {
				t.Errorf("len(keys) = %d ==> len(k) = %d, want %d", len(keys), len(k), tt.expectedParts)
			}
		})
	}
}

func TestMaxParts(t *testing.T) {
	tests := []struct {
		size          int
		expectedParts int
	}{
		{size: 10, expectedParts: 1},
		{size: 63, expectedParts: 1},
		{size: 64, expectedParts: 1},
		{size: 65, expectedParts: 2},
		{size: 100, expectedParts: 2},
		{size: 127, expectedParts: 2},
		{size: 128, expectedParts: 2},
		{size: 129, expectedParts: 3},
		{size: 191, expectedParts: 3},
		{size: 192, expectedParts: 3},
		{size: 193, expectedParts: 4},
		{size: 255, expectedParts: 4},
		{size: 256, expectedParts: 4},
		{size: 257, expectedParts: 5},
		{size: 1023, expectedParts: 16},
		{size: 1024, expectedParts: 16},
		{size: 1025, expectedParts: 17},
		{size: 10_000, expectedParts: 157},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("size=%d,parts=%d", tt.size, tt.expectedParts), func(t *testing.T) {
			p := maxP(tt.size)
			if p != tt.expectedParts {
				t.Errorf("sizeToParts(%d)=%d, want %d", tt.size, p, tt.expectedParts)
			}
		})
	}
}

func maxP(size int) int {
	if size%64 != 0 {
		return size/64 + 1
	}
	return size / 64
}

func generateKeys(size, seed int) []uint64 {
	keys := make([]uint64, size)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := range keys {
		keys[i] = r.Uint64()
	}
	return keys
}
