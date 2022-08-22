package io

// Consumer can receive an instance of A and perform some operation on it.
type Consumer[A any] func(A) IOUnit

// CoMap changes the input argument of the consumer.
func CoMap[A any, B any](ca Consumer[A], f func(b B) A) Consumer[B] {
	return func(b B) IOUnit {
		return Delay(func() IOUnit {
			return ca(f(b))
		})
	}
}
