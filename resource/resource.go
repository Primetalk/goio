// Package resource provides some means to deal with resources.
package resource

import (
	"log"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Closable is a value that is accompanied with the Close().
// This is an internal structure that should not be used outside of the resource package.
type Closable[A any] struct {
	Value A
	Close func() io.IO[fun.Unit]
}

// Resource[A] is an structure that can only be _used_ via Use.
// Unfortunately, it's not an interface, because interface methods do not support generics
// at the moment.
type Resource[A any] io.IO[Closable[A]]

// Use is a only way to access the resource instance.
// It guarantees that the resource instance will be closed after use
// regardless of the failure/success result.
func Use[A any, B any](res Resource[A], f func(A) io.IO[B]) io.IO[B] {
	return io.FlatMap(io.IO[Closable[A]](res), func(cl Closable[A]) io.IO[B] {
		iob := f(cl.Value)
		return io.Fold(iob,
			func(b B) io.IO[B] {
				return io.Map(cl.Close(), func(fun.Unit) B {
					return b
				})
			},
			func(err error) io.IO[B] {
				iocl := cl.Close()
				ioclSafe := io.Recover(iocl, func(err2 error) io.IO[fun.Unit] {
					log.Printf("double error during resource release: %+v", err2)
					return io.IOUnit1
				})
				return io.FlatMap(ioclSafe, func(fun.Unit) io.IO[B] {
					return io.Fail[B](err)
				})
			})
	})
}

// NewResource constructs a resource given two functions - acquire and release.
func NewResource[A any](acquire io.IO[A], release func(A) io.IO[fun.Unit]) Resource[A] {
	return Resource[A](io.Map(acquire, func(a A) Closable[A] {
		return Closable[A]{
			Value: a,
			Close: func() io.IO[fun.Unit] {
				return release(a)
			},
		}
	}))
}

// NewResourceFromIOClosable - is an internal function that constructs a resource from closable IO.
func NewResourceFromIOClosable[A any](cl io.IO[Closable[A]]) Resource[A] {
	return Resource[A](cl)
}

// ClosableMap is an internal function to map closable using the provided function.
func ClosableMap[A any, B any](ra Closable[A], f func(a A) B) Closable[B] {
	return Closable[B]{
		Value: f(ra.Value),
		Close: ra.Close,
	}
}

// ClosableFlatMap flatmaps the closable. Allows to construct have more than one resource in scope.
func ClosableFlatMap[A any, B any](ca Closable[A], f func(a A) Closable[B]) Closable[B] {
	cb := f(ca.Value)
	return Closable[B]{
		Value: cb.Value,
		Close: func() io.IO[fun.Unit] {
			return io.Fold(cb.Close(),
				func(fun.Unit) io.IO[fun.Unit] {
					return ca.Close()
				},
				func(err2 error) io.IO[fun.Unit] {
					log.Printf("double error during closable release: %+v", err2)
					return ca.Close()
				},
			)
		},
	}
}

// Map maps the resource value using the provided conversion function.
func Map[A any, B any](ra Resource[A], f func(a A) B) Resource[B] {
	return Resource[B](io.Map(io.IO[Closable[A]](ra),
		func(ca Closable[A]) Closable[B] {
			return ClosableMap(ca, f)
		}))
}

// FlatMap allows to add another resource to scope. Both will be released in reverse order.
func FlatMap[A any, B any](ra Resource[A], f func(a A) Resource[B]) Resource[B] {
	return Resource[B](io.FlatMap(io.IO[Closable[A]](ra),
		func(ca Closable[A]) io.IO[Closable[B]] {
			cb := ClosableMap(ca, func(a A) io.IO[Closable[B]] { return io.IO[Closable[B]](f(a)) })
			return ClosableIOTransform(cb)
		}))
}

// ClosableIOTransform transforms a closable of io closable to just io closable.
func ClosableIOTransform[A any](cioca Closable[io.IO[Closable[A]]]) (ioca io.IO[Closable[A]]) {
	return io.Eval(func() (ca Closable[A], err error) {
		defer fun.RecoverToErrorVar("resource.ClosableIOTransform", &err)
		ca = ClosableFlatMap(cioca, func(ioca io.IO[Closable[A]]) (ca1 Closable[A]) {
			ca1, err = io.UnsafeRunSync(ioca)
			return
		})
		return
	})
}

// UnbufferedChannel returns a resource that manages a channel.
func UnbufferedChannel[A any]() Resource[chan A] {
	return NewResource(io.MakeUnbufferedChannel[A](), func(ch chan A) io.IOUnit {
		return io.CloseChannel(ch)
	})
}

// Fail creates a resource that will fail during acquisition.
func Fail[A any](err error) Resource[A] {
	return NewResource(io.Fail[A](err), fun.Const[A](io.IOUnit1))
}
