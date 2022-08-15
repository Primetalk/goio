// Package option contains Option[A] implementation.
package option

import "github.com/primetalk/goio/fun"

// Option[A] can represent a value or an absent value of type A.
type Option[A any] struct {
	ValueOrNil *A
}

// None constructs an option without value.
func None[A any]() Option[A] {
	return Option[A]{}
}

// Some constructs an option with value.
func Some[A any](a A) Option[A] {
	return Option[A]{
		ValueOrNil: &a,
	}
}

// Map applies a function to the value inside option if any.
func Map[A any, B any](oa Option[A], f func(A) B) Option[B] {
	return Fold(oa,
		func(a A) Option[B] {
			return Some(f(a))
		},
		None[B],
	)
}

// Fold transforms all possible values of OptionA using two provided functions.
func Fold[A any, B any](oa Option[A], f func(A) B, g func() B) (b B) {
	if oa.ValueOrNil == nil {
		b = g()
	} else {
		b = f(*oa.ValueOrNil)
	}
	return
}

// Filter leaves the value inside option only if predicate is true.
func Filter[A any](oa Option[A], predicate func(A) bool) Option[A] {
	return Fold(oa,
		func(a A) Option[A] {
			if predicate(a) {
				return oa
			} else {
				return None[A]()
			}
		},
		None[A],
	)
}

// FlatMap converts an internal value if it is present using the provided function.
func FlatMap[A any, B any](oa Option[A], f func(A) Option[B]) Option[B] {
	return Fold(oa,
		func(a A) Option[B] {
			return f(a)
		},
		None[B],
	)
}

// Flatten simplifies option of option to just Option[A].
func Flatten[A any](ooa Option[Option[A]]) Option[A] {
	return FlatMap(ooa, fun.Identity[Option[A]])
}

// Get is an unsafe function that unwraps the value from the option.
func Get[A any](oa Option[A]) A {
	return Fold(oa, fun.Identity[A], fun.Nothing[A])
}

// ForEach runs the given function on the value if it's available.
func ForEach[A any](oa Option[A], f func(A)) {
	if oa.ValueOrNil != nil {
		f(*oa.ValueOrNil)
	}
}

// IsDefined checks whether the option contains a value.
func IsDefined[A any](oa Option[A]) bool {
	return oa.ValueOrNil != nil
}

// IsEmpty checks whether the option is empty.
func IsEmpty[A any](oa Option[A]) bool {
	return oa.ValueOrNil == nil
}
