// Package fun provides reusable general-purpose functions (Const, Swap, Curry) and data structures (Unit, Either).
package fun

// Const creates a function that will ignore it's input and return the specified value.
func Const[A any, B any](b B) func(A) B {
	return func(A) B {
		return b
	}
}

// ConstUnit creates a function that will ignore it's Unit input and return the specified value.
func ConstUnit[B any](b B) func(Unit) B {
	return Const[Unit](b)
}

// Swap returns a curried function with swapped order of arguments.
func Swap[A any, B any, C any](f func(a A) func(b B) C) func(b B) func(a A) C {
	return func(b B) func(a A) C {
		return func(a A) C {
			return f(a)(b)
		}
	}
}

// Curry takes a function that has two arguments and returns a function with two argument lists.
func Curry[A any, B any, C any](f func(a A, b B) C) func(a A) func(b B) C {
	return func(a A) func(b B) C {
		return func(b B) C {
			return f(a, b)
		}
	}
}
