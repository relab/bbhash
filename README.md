# BBHash

BBHash is a fast Go implementation of a minimal perfect hash function for large key sets.

## Installing the module for use in your own project

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
	bb, err := bbhash.New(keys, bbhash.Gamma(1.5))
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

## Advanced usage

The `bbhash.New` function takes a slice of keys as its first argument.
The keys must be unique and of type `uint64`.
`New` also takes zero or more `bbhash.Option` arguments.
These are the available options:

| Option | Description |
| --- | --- |
| `Gamma(float64)`     | Set the gamma parameter of the BBHash algorithm. Default is 2.0.               |
| `InitialLevels(int)` | Set the initial number of levels in the BBHash algorithm. Default is 32.       |
| `Partitions(int)`    | Set the number of partitions to split the keys into and compute parallel.      |
| `WithReverseMap()`   | Create a reverse map that allows you to retrieve the key from the hash index.  |
| `Parallel()`         | Use parallelism in the BBHash algorithm. Prefer the Partitions option instead. |

The options can be combined like this:

```go
bb, err := bbhash.New(keys)
bb, err := bbhash.New(keys, bbhash.Gamma(1.5), bbhash.InitialLevels(64))
bb, err := bbhash.New(keys, bbhash.InitialLevels(20))
bb, err := bbhash.New(keys, bbhash.Gamma(1.5), bbhash.Partitions(4))
bb, err := bbhash.New(keys, bbhash.Gamma(1.5), bbhash.Partitions(4), bbhash.WithReverseMap())
```

But the following combinations are not supported:

```go
bb, err := bbhash.New(keys, bbhash.Parallel(), bbhash.Partitions(4))
bb, err := bbhash.New(keys, bbhash.Parallel(), bbhash.WithReverseMap())
```

## Credits

Implemented by Hein Meling.

The implementation is mainly inspired by Sudhi Herle's [Go implementation](https://github.com/opencoff/go-bbhash).
Damian Gryski also has a Go [implementation](<https://github.com/dgryski/go-boomphf>).

The algorithm is described in the paper:
[Fast and Scalable Minimal Perfect Hashing for Massive Key Sets](https://arxiv.org/abs/1702.03154)
Antoine Limasset, Guillaume Rizk, Rayan Chikhi, and Pierre Peterlongo.
Their C++ implementation is available [here](https://github.com/rizkg/BBHash.).
