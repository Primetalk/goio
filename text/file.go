package text

import (
	"os"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
)

// ReadOnlyFile returns a resource for the specified filename.
func ReadOnlyFile(name string) resource.Resource[*os.File] {
	return resource.NewResource(
		io.Eval( func () (*os.File, error) {
			return os.Open(name)
		}),
		func(f *os.File) io.IO[fun.Unit] {
			return io.FromUnit(func() (error) {
				return f.Close()
			})
		},
	)
}
