//go:build !amd64 && !arm64

package fast

func Hash64(seed uint64, buf []byte) uint64 {
	return hash64(seed, buf)
}
