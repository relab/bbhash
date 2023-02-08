# BBHash

BBHash is a fast Go implementation of a minimal perfect hash function for large key sets.

## Usage

```go
import "github.com/relab/bbhash"
```

## Credits

The algorithm is described in the paper:
[Fast and Scalable Minimal Perfect Hashing for Massive Key Sets](https://arxiv.org/abs/1702.03154)
Antoine Limasset, Guillaume Rizk, Rayan Chikhi, and Pierre Peterlongo.
Their C++ implementation is available [here](https://github.com/rizkg/BBHash.).

The implementation is mainly inspired by Sudhi Herle's Go implementation of [BBHash](https://github.com/opencoff/go-bbhash).
Damian Gryski also has a Go [implementation](<https://github.com/dgryski/go-boomphf>).
