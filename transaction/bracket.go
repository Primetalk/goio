package transaction

import (
	"fmt"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Bracket executes user computation with transactional guarantee.
// If user computation is successful - commit is executed.
// Otherwise - rollback.
func Bracket[A any, T any](acquire io.IO[T], commit func(T) io.IOUnit, rollback func(T) io.IOUnit) func(tr func(t T) io.IO[A]) io.IO[A] {
	return func(tr func(t T) io.IO[A]) io.IO[A] {
		return io.FlatMap(acquire, func(t T) io.IO[A] {
			return io.Fold(tr(t),
				func(a A) io.IO[A] {
					return io.MapConst(commit(t), a)
				},
				func(err error) io.IO[A] {
					return io.Fold(rollback(t),
						func(u fun.Unit) io.IO[A] {
							return io.Fail[A](err)
						},
						func(err2 error) io.IO[A] {
							return io.AndThen(
								io.FromPureEffect(func() {
									fmt.Printf("duplicated error in TransactionalBracket: %+v", err2)
								}),
								io.Fail[A](err),
							)
						},
					)
				},
			)
		})
	}
}
