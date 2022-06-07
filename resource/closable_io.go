package resource

import "github.com/primetalk/goio/io"

// ClosableIO is a simple resource that implements Close method.
type ClosableIO interface {
	Close() io.IOUnit
}

// FromClosableIO constructs a new resource from some value that
// itself supports method Close.
func FromClosableIO[A ClosableIO](ioa io.IO[A]) Resource[A] {
	return NewResource(ioa, func(a A) io.IOUnit { return a.Close() })
}

// BoundedExecutionContextResource returns a resource that is a bounded execution context.
func BoundedExecutionContextResource(size int64, queueLimit int) Resource[io.ExecutionContext] {
	return FromClosableIO(io.Pure(func() io.ExecutionContext {
		return io.BoundedExecutionContext(size, queueLimit)
	}))
}
