# BBHash

BBHash is a fast Go implementation of a minimal perfect hash function for large key sets.

## Installing the package for use in your own project

```sh
% go get github.com/relab/bbhash
```

## How to use the package

```go
import (
	"fmt"

	"github.com/relab/bbhash"
)

func ExampleBBHash_Find() {
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bb, err := bbhash.NewSequential(1.5, 0, keys)
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		hashIndex := bb.Find(key)
		fmt.Printf("%d, ", hashIndex)
	}
	fmt.Println()
	// Output:
	// 2, 6, 8, 3, 5, 7, 1, 9, 10, 4,
}
```

## Credits

Implemented by Hein Meling.

The implementation is mainly inspired by Sudhi Herle's [Go implementation](https://github.com/opencoff/go-bbhash).
Damian Gryski also has a Go [implementation](<https://github.com/dgryski/go-boomphf>).

The algorithm is described in the paper:
[Fast and Scalable Minimal Perfect Hashing for Massive Key Sets](https://arxiv.org/abs/1702.03154)
Antoine Limasset, Guillaume Rizk, Rayan Chikhi, and Pierre Peterlongo.
Their C++ implementation is available [here](https://github.com/rizkg/BBHash.).
