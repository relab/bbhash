package bbhash

const (
	// defaultGamma is the default expansion factor for the bit vector.
	defaultGamma = 2.0

	// minimalGamma is the smallest allowed expansion factor for the bit vector.
	minimalGamma = 0.5

	// Heuristic: 32 levels should be enough for even very large key sets
	initialLevels = 32

	// Maximum number of attempts (level) at making a perfect hash function.
	// Per the paper, each successive level exponentially reduces the
	// probability of collision.
	maxLevel = 255

	// Maximum number of partitions.
	maxPartitions = 255
)

type options struct {
	gamma         float64
	initialLevels int
	partitions    int
	parallel      bool
	reverseMap    bool
}

func newOptions(opts ...Options) *options {
	o := &options{
		gamma:         defaultGamma,
		initialLevels: initialLevels,
		partitions:    1,
		parallel:      false,
		reverseMap:    false,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Options func(*options)

// Gamma sets the gamma parameter for creating a BBHash.
// The gamma parameter is the expansion factor for the bit vector; the paper recommends
// a value of 2.0. The larger the value the more space will be consumed by the BBHash.
func Gamma(gamma float64) Options {
	return func(o *options) {
		o.gamma = max(gamma, minimalGamma)
	}
}

// InitialLevels sets the initial number of levels to use when creating a BBHash.
func InitialLevels(levels int) Options {
	return func(o *options) {
		o.initialLevels = levels
	}
}

// Partitions sets the number of partitions to use when creating a BBHash.
// The keys are partitioned into the given the number partitions.
// Setting partitions to less than 2 results in a single BBHash, wrapped in a BBHash.
func Partitions(partitions int) Options {
	return func(o *options) {
		o.partitions = max(min(partitions, maxPartitions), 1)
	}
}

// Parallel creates a BBHash by sharding the keys across multiple goroutines.
// This option is not compatible with the Partitions option.
func Parallel() Options {
	return func(o *options) {
		o.parallel = true
	}
}

// WithReverseMap creates a reverse map when creating a BBHash.
func WithReverseMap() Options {
	return func(o *options) {
		o.reverseMap = true
	}
}
