package fun

// Pair is a data structure that has two values.
type Pair[A any, B any] struct {
	V1 A
	V2 B
}

// NewPair constructs the pair.
func NewPair[A any, B any](a A, b B) Pair[A, B] { return Pair[A, B]{V1: a, V2: b} }

// PairV1 returns the first element of the pair.
func PairV1[A any, B any](p Pair[A, B]) A { return p.V1 }

// PairV2 returns the second element of the pair.
func PairV2[A any, B any](p Pair[A, B]) B { return p.V2 }
