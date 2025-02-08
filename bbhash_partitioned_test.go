package bbhash_test

import (
	"testing"

	"github.com/relab/bbhash"
	"github.com/relab/bbhash/internal/test"
)

func TestNewOptions(t *testing.T) {
	const (
		small = 100  // too few keys to enable partitioning
		limit = 1000 // at least limit keys are needed to enable partitioning
	)
	tests := []struct {
		name           string
		size           int
		opts           []bbhash.Options
		wantPanic      bool
		wantPartitions int // defaults to 1 if not set
	}{
		{name: "sequential", size: 0, opts: []bbhash.Options{}, wantPanic: true},
		{name: "sequential", size: 1, opts: []bbhash.Options{}, wantPanic: false},
		{name: "sequential", size: small, opts: []bbhash.Options{}, wantPanic: false},
		{name: "sequential", size: limit, opts: []bbhash.Options{}, wantPanic: false},
		{name: "parallel", size: small, opts: []bbhash.Options{bbhash.Parallel()}, wantPanic: false},
		{name: "parallel", size: limit, opts: []bbhash.Options{bbhash.Parallel()}, wantPanic: false},

		{name: "partitions=-1", size: small, opts: []bbhash.Options{bbhash.Partitions(-1)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions=-1", size: limit, opts: []bbhash.Options{bbhash.Partitions(-1)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions=0", size: small, opts: []bbhash.Options{bbhash.Partitions(0)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions=0", size: limit, opts: []bbhash.Options{bbhash.Partitions(0)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions=1", size: small, opts: []bbhash.Options{bbhash.Partitions(1)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions=1", size: limit, opts: []bbhash.Options{bbhash.Partitions(1)}, wantPanic: false, wantPartitions: 1},

		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(2)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(2)}, wantPanic: false, wantPartitions: 2},
		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(3)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(3)}, wantPanic: false, wantPartitions: 3},
		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(5)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(5)}, wantPanic: false, wantPartitions: 5},
		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(20)}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(20)}, wantPanic: false, wantPartitions: 20},

		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(1), bbhash.Parallel()}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(1), bbhash.Parallel()}, wantPanic: false, wantPartitions: 1},
		{name: "partitions", size: small, opts: []bbhash.Options{bbhash.Partitions(2), bbhash.Parallel()}, wantPanic: true},
		{name: "partitions", size: limit, opts: []bbhash.Options{bbhash.Partitions(2), bbhash.Parallel()}, wantPanic: true},

		{name: "reversemap", size: small, opts: []bbhash.Options{bbhash.WithReverseMap()}, wantPanic: false},
		{name: "reversemap", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap()}, wantPanic: false},

		{name: "reversemap/parallel", size: small, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Parallel()}, wantPanic: true},
		{name: "reversemap/parallel", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Parallel()}, wantPanic: true},

		{name: "reversemap/partitions", size: small, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(2)}, wantPanic: false, wantPartitions: 1},
		{name: "reversemap/partitions", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(2)}, wantPanic: false, wantPartitions: 2},
		{name: "reversemap/partitions", size: small, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(3)}, wantPanic: false, wantPartitions: 1},
		{name: "reversemap/partitions", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(3)}, wantPanic: false, wantPartitions: 3},
		{name: "reversemap/partitions", size: small, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(5)}, wantPanic: false, wantPartitions: 1},
		{name: "reversemap/partitions", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(5)}, wantPanic: false, wantPartitions: 5},
		{name: "reversemap/partitions", size: small, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(20)}, wantPanic: false, wantPartitions: 1},
		{name: "reversemap/partitions", size: limit, opts: []bbhash.Options{bbhash.WithReverseMap(), bbhash.Partitions(20)}, wantPanic: false, wantPartitions: 20},
	}
	for _, tt := range tests {
		keys := generateKeys(tt.size, 123)
		wantPartitions := max(tt.wantPartitions, 1)
		t.Run(test.Name(tt.name, []string{"keys", "partitions"}, tt.size, wantPartitions), func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("test %s did not panic", tt.name)
					}
				}()
			}
			bb, err := bbhash.New(keys, tt.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if bb == nil {
				t.Fatal("bb is nil")
			}
			if bb.Partitions() != wantPartitions {
				t.Errorf("got %d partitions, want %d", bb.Partitions(), wantPartitions)
			}
		})
	}
}

// TODO add tests for the BBHash2.Find method
// TODO add tests for the BBHash2.Key method when both reverse map and partitioning is used
