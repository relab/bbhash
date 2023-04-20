package bbhash

func split(keys []uint64, parts int) [][]uint64 {
	n := len(keys)
	z := n / parts
	r := n % parts
	k := make([][]uint64, parts)
	for j := 0; j < parts; j++ {
		x := z * j
		y := x + z
		if j == parts-1 {
			y += r
		}
		k[j] = keys[x:y]
	}
	return k
}

func splitX(keys []uint64) [][]uint64 {
	n := len(keys)
	const w = 512
	parts := (n + w - 1) / w
	z := n / parts
	r := n % parts
	k := make([][]uint64, parts)
	for j := 0; j < parts; j++ {
		x := z * j
		y := x + z
		if j == parts-1 {
			y += r
		}
		k[j] = keys[x:y]
	}
	return k
}

func splitP(keys []uint64, parts int) [][]uint64 {
	n := len(keys)
	// todo can use words() func here
	maxParts := n / 64
	if n%64 > 0 {
		maxParts++
	}
	if parts > maxParts {
		parts = maxParts
	}
	z := n / parts
	r := n % parts
	k := make([][]uint64, parts)
	for j := 0; j < parts; j++ {
		x := z * j
		y := x + z
		if j == parts-1 {
			y += r
		}
		k[j] = keys[x:y]
	}
	return k
}
