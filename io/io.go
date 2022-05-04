package io

type IO[A any] interface {
	unsafeRun() (A, error)
}


func UnsafeRunSync[A any](io IO[A]) (A, error) {
	return io.unsafeRun()
}

func Eval[A any](f func () (A, error)) IO[A] {
	return evalImpl[A]{
		f: f,
	}
}


type evalImpl[A any] struct {
	f func () (A, error)
}

func (e evalImpl[A])unsafeRun() (res A, err error) {
	defer RecoverToErrorVar("Eval.unsafeRun", &err)
	return e.f()
}

func Map[A any, B any](ioA IO[A], f func(a A)(B, error)) IO[B] {
	return mapImpl[A, B]{
		io: ioA,
		f: f,
	}
}

func MapPure[A any, B any](ioA IO[A], f func(a A)B) IO[B] {
	return mapImpl[A, B]{
		io: ioA,
		f: func (a A)(B, error){ return f(a), nil},
	}
}

type mapImpl[A any, B any] struct {
	io IO[A]
	f func(a A)(B, error)
}

func (e mapImpl[A, B])unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("Map.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a)
	}
	return
}

func FlatMap[A any, B any](ioA IO[A], f func(a A) IO[B]) IO[B] {
	return flatMapImpl[A, B]{
		io: ioA,
		f: f,
	}
}

type flatMapImpl[A any, B any] struct {
	io IO[A]
	f func(a A) IO[B]
}

func (e flatMapImpl[A, B])unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("FlatMap.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a).unsafeRun()
	}
	return
}

func Lift[A any](a A) IO[A] {
	return Eval(func()(A, error){return a, nil})
}

func Fail[A any](err error) IO[A] {
	return Eval(func()(a A, err1 error){
		err1 = err
		return 
	})
}
