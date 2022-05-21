package fun

// Either is a simple data structure that can have either left value or right value
type Either[A any, B any] struct {
	IsLeft bool
	Left   A
	Right  B
}

func Left[A any, B any](a A) Either[A, B] {
	return Either[A, B]{
		IsLeft: true,
		Left:   a,
	}
}

func Right[A any, B any](b B) Either[A, B] {
	return Either[A, B]{
		IsLeft: false,
		Right:  b,
	}
}

func IsLeft[A any, B any](eab Either[A, B]) bool {
	return eab.IsLeft
}

func IsRight[A any, B any](eab Either[A, B]) bool {
	return !eab.IsLeft
}

// Fold pattern matches Either with two given pattern match handlers
func Fold[A any, B any, C any](eab Either[A, B], left func(A)C, right func(B)C) C {
	if eab.IsLeft {
		return left(eab.Left)
	} else {
		return right(eab.Right)
	}
}
