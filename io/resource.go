package io

type Resource[A any] struct {
	Aquire IO[A]
	Release func(A)IO[Unit]
}

func NewResource[A any](aquire IO[A], release func(A)IO[Unit]) Resource[A] {
	return Resource[A]{
		Aquire: aquire,
		Release: release,
	}
}

func Use[A any, B any](res Resource[A], f func (A) IO[B]) IO[B] {
	aquire, release := res.Aquire, res.Release
	
	ioab := FlatMap(aquire, func (a A) IO[Pair[A, B]] {
		iob := f(a)
		return Map(iob, func(b B)Pair[A, B]{ return NewPair(a, b) })
	})

	iob := FlatMap(ioab, func (ab Pair[A, B]) IO[B] {
		iou := release(ab.A)
		return Map(iou, func (Unit) B { return ab.B})
	})
	return iob
}
