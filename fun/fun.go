package fun

// Const creates a function that will ignore it's input and return the specified value
func Const[A any, B any](b B)func(A)B {
	return func(A)B {
		return b
	}
}

func ConstUnit[B any](b B) func(Unit)B {
	return Const[Unit](b)
}

func Swap[A any, B any, C any](f func(a A)func(b B)C) func(b B)func(a A)C {
	return func(b B)func(a A)C {
		return func(a A)C {
			return f(a)(b)
		}
	}
}

func Curry[A any, B any, C any](f func(a A, b B)C) func(a A)func(b B)C {
	return func(a A)func(b B)C {
		return func(b B) C {
			return f(a, b)
		}
	}
}

