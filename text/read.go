// Package text provides some utilities to work with text files.
package text

import (
	"errors"
	fio "io"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

// ReadByteChunks reads chunks from the reader.
func ReadByteChunks(reader fio.Reader, chunkSize int) stream.Stream[[]byte] {
	return stream.Stream[[]byte](io.Eval(func() (res stream.StepResult[[]byte], err error) {
		bytes := make([]byte, chunkSize)
		var cnt int
		cnt, err = reader.Read(bytes)
		if err == fio.EOF {
			err = nil
			var cont stream.Stream[[]byte]
			if cnt == 0 {
				cont = stream.Empty[[]byte]()
			} else {
				cont = stream.Lift(bytes[0:cnt])
			}
			res = stream.NewStepResultEmpty(cont)
		} else if err == nil {
			if cnt == 0 {
				res = stream.NewStepResultEmpty(stream.Empty[[]byte]())
			} else {
				res = stream.NewStepResult(bytes[0:cnt], ReadByteChunks(reader, chunkSize))
			}
		}
		return
	}))
}

var emptyByteChunkStream = stream.Empty[[]byte]()

// SplitBySeparator splits byte-chunk stream by the given separator.
func SplitBySeparator(stm stream.Stream[[]byte], sep byte, shouldReturnLastIncompleteLine bool) stream.Stream[[]byte] {
	return stream.StateFlatMapWithFinish(stm, []byte{},
		func(a []byte, state []byte) io.IO[fun.Pair[[]byte, stream.Stream[[]byte]]] {
			return io.Pure(func() fun.Pair[[]byte, stream.Stream[[]byte]] {
				var resultState []byte
				var stm stream.Stream[[]byte]
				parts := splitBy(sep, a, [][]byte{})
				if len(parts) == 0 {
					stm = stream.Fail[[]byte](errors.New("unexpected len==0 from splitBy"))
				} else if len(parts) == 1 {
					// separator not found
					resultState = append(state, a...)
					stm = emptyByteChunkStream
				} else {
					parts[0] = append(state, parts[0]...)
					resultState = parts[len(parts)-1]
					stm = stream.LiftMany(parts[0 : len(parts)-1]...)
				}
				return fun.NewPair(resultState, stm)
			})

		},
		func(s []byte) stream.Stream[[]byte] {
			if len(s) > 0 && shouldReturnLastIncompleteLine {
				return stream.Lift(s)
			} else {
				return emptyByteChunkStream
			}
		})
}

func indexOf[A comparable](element A, data []A) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

// splitBy returns at least one part when separator is not found.
// Even if len(data) == 0.
func splitBy[A comparable](sep A, data []A, prefixParts [][]A) (parts [][]A) {
	i := indexOf(sep, data)
	if i == -1 {
		return append(prefixParts, data)
	} else {
		return splitBy(sep, data[i+1:], append(prefixParts, data[:i]))
	}
}

// MapToStrings converts stream of bytes to strings.
func MapToStrings(stm stream.Stream[[]byte]) stream.Stream[string] {
	return stream.Map(stm, func(a []byte) string { return string(a) })
}

const DefaultChunkSize = 4096

// ReadLines reads text file line-by-line.
// If there is a last line that is not terminated by '\n', it is ignored.
func ReadLines(reader fio.Reader) stream.Stream[string] {
	chunks := ReadByteChunks(reader, DefaultChunkSize)
	rows := SplitBySeparator(chunks, '\n', false)
	return MapToStrings(rows)
}

// ReadLinesWithLastNonFinishedLine reads text file line-by-line.
// Also returns the last line that is not terminated by '\n'
func ReadLinesWithNonFinishedLine(reader fio.Reader) stream.Stream[string] {
	chunks := ReadByteChunks(reader, DefaultChunkSize)
	rows := SplitBySeparator(chunks, '\n', true)
	return MapToStrings(rows)
}
