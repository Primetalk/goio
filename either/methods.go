package either

// Fold transforms either side of e into a common result type.
func (e Either[A, B]) Fold[C any](left func(A) C, right func(B) C) C {
	return Fold(e, left, right)
}
