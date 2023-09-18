package bbhash_test

import (
	"fmt"

	"github.com/relab/bbhash"
)

func ExampleBBHash_Find() {
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bb, err := bbhash.NewSequential(1.5, keys)
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		hashIndex := bb.Find(key)
		fmt.Printf("%d, ", hashIndex)
	}
	fmt.Println()
	// Output:
	// 2, 6, 1, 4, 9, 3, 8, 5, 10, 7,
}
