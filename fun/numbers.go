package fun

import "golang.org/x/exp/constraints"

// Number is a generic number interface that covers all Go number types.
type Number interface {
	constraints.Integer |
		constraints.Float |
		constraints.Complex
}

// Min - returns the minimum value.
// See https://go.dev/blog/intro-generics
func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// Max - returns the maximum value.
// See https://go.dev/blog/intro-generics
func Max[T constraints.Ordered](x, y T) T {
	if x < y {
		return y
	}
	return x
}
