package resource

import (
	"log"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

type Closable[A any] struct {
	Value A
	Close func() io.IO[fun.Unit]
}

type Resource[A any] io.IO[Closable[A]]

func Use[A any, B any](res Resource[A], f func(A) io.IO[B]) io.IO[B] {
	return io.FlatMap(res.(io.IO[Closable[A]]), func (cl Closable[A]) io.IO[B] { 
		iob := f(cl.Value)
		return io.Fold(iob, 
			func (b B) io.IO[B] {
				return io.Map(cl.Close(), func (fun.Unit) B {
					return b
				})
			}, 
			func (err error) io.IO[B] {
				iocl := cl.Close()
				ioclSafe := io.Recover(iocl, func (err2 error)io.IO[fun.Unit]{
					log.Printf("double error during resource release: %+v", err2)
					return io.IOUnit1
				})
				return io.FlatMap(ioclSafe, func (fun.Unit) io.IO[B] {
					return io.Fail[B](err)
				})
			})
	})
}

func NewResource[A any](acquire io.IO[A], release func(A)io.IO[fun.Unit]) Resource[A] {
	return io.Map(acquire, func (a A) Closable[A] {
		return Closable[A]{
			Value: a,
			Close: func() io.IO[fun.Unit] {
				return release(a)
			},
		}
	}) 
}

func NewResourceFromIOClosable[A any](cl io.IO[Closable[A]]) Resource[A] {
	return cl
}

func ClosableMap[A any, B any](ra Closable[A], f func (a A) B ) Closable[B] {
	return Closable[B] {
		Value: f(ra.Value),
		Close: ra.Close,
	}
}

func ClosableFlatMap[A any, B any](ca Closable[A], f func (a A) Closable[B] ) Closable[B] {
	cb := f(ca.Value)
	return Closable[B] {
		Value: cb.Value,
		Close: func () io.IO[fun.Unit] {
			return io.Fold(cb.Close(), 
			func (fun.Unit) io.IO[fun.Unit] {
				return ca.Close()
			},
			func (err2 error) io.IO[fun.Unit] {
				log.Printf("double error during closable release: %+v", err2)
				return ca.Close()
			},
			)
		} ,
	}
}

func Map[A any, B any](ra Resource[A], f func (a A) B ) Resource[B] {
	return io.Map[Closable[A]](ra, 
		func (ca Closable[A]) Closable[B] { 
			return ClosableMap(ca, f) 
		})
}

func FlatMap[A any, B any](ra Resource[A], f func (a A) Resource[B] ) Resource[B] {
	return io.FlatMap[Closable[A]](ra, 
		func (ca Closable[A]) io.IO[Closable[B]] {
			cb := ClosableMap(ca, func (a A) io.IO[Closable[B]] { return f(a)})
			return ClosableIOTransform(cb)
		})
}

func ClosableIOTransform[A any](cioca Closable[io.IO[Closable[A]]]) (ioca io.IO[Closable[A]]) {
	return io.Eval(func () (ca Closable[A], err error) {
		defer io.RecoverToErrorVar("resource.ClosableIOTransform", &err)
		ca = ClosableFlatMap(cioca, func (ioca io.IO[Closable[A]]) (ca1 Closable[A]) {
			ca1, err = io.UnsafeRunSync(ioca)
			return
		})
		return
	}) 
}
