package io

// Map transforms the successful result of ioa without executing it eagerly.
func (ioa IO[A]) Map[B any](f func(A) B) IO[B] {
	return Map(ioa, f)
}

// FlatMap composes ioa with another lazy computation.
func (ioa IO[A]) FlatMap[B any](f func(A) IO[B]) IO[B] {
	return FlatMap(ioa, f)
}

// Fold handles both the successful and failed outcomes of ioa.
func (ioa IO[A]) Fold[B any](success func(A) IO[B], failure func(error) IO[B]) IO[B] {
	return Fold(ioa, success, failure)
}
