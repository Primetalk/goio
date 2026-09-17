package stream

// Map transforms every value emitted by stm.
func (stm Stream[A]) Map[B any](f func(A) B) Stream[B] {
	return Map(stm, f)
}

// FlatMap replaces every value emitted by stm with another stream.
func (stm Stream[A]) FlatMap[B any](f func(A) Stream[B]) Stream[B] {
	return FlatMap(stm, f)
}

// Through applies pipe to stm.
func (stm Stream[A]) Through[B any](pipe Pipe[A, B]) Stream[B] {
	return Through(stm, pipe)
}
