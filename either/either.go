package either

// Either is a simple data structure that can have either left value or right value.
type Either[A any, B any] struct {
	IsLeft bool
	Left   A
	Right  B
}

// Left constructs Either that is left.
func Left[A any, B any](a A) Either[A, B] {
	return Either[A, B]{
		IsLeft: true,
		Left:   a,
	}
}

// Right constructs Either that is right.
func Right[A any, B any](b B) Either[A, B] {
	return Either[A, B]{
		IsLeft: false,
		Right:  b,
	}
}

// IsLeft checks whether the provided Either is left or not.
func IsLeft[A any, B any](eab Either[A, B]) bool {
	return eab.IsLeft
}

// IsRight checks whether the provided Either is right or not.
func IsRight[A any, B any](eab Either[A, B]) bool {
	return !eab.IsLeft
}

// Fold pattern matches Either with two given pattern match handlers
func Fold[A any, B any, C any](eab Either[A, B], left func(A) C, right func(B) C) C {
	if eab.IsLeft {
		return left(eab.Left)
	} else {
		return right(eab.Right)
	}
}
