package stream

// AddSeparatorAfterEachElement adds a separator after each stream element
func AddSeparatorAfterEachElement[A any](stm Stream[A], sep A) Stream[A] {
	return FlatMap(stm, func(a A) Stream[A] {
		return LiftMany(a, sep)
	})
}
