package option

// Map transforms the contained value when this option is defined.
func (oa Option[A]) Map[B any](f func(A) B) Option[B] {
	return Map(oa, f)
}

// FlatMap composes this option with another optional computation.
func (oa Option[A]) FlatMap[B any](f func(A) Option[B]) Option[B] {
	return FlatMap(oa, f)
}

// Match transforms either the contained value or the empty case.
func (oa Option[A]) Match[B any](defined func(A) B, empty func() B) B {
	return Match(oa, defined, empty)
}
