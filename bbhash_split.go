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
