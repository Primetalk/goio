// Package text provides some utilities to work with text files.
package text

import (
	fio "io"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

// ReadByteChunks reads chunks from the reader.
func ReadByteChunks(reader fio.Reader, chunkSize int) stream.Stream[[]byte] {
	return io.Eval(func() (res stream.StepResult[[]byte], err error) {
		bytes := make([]byte, chunkSize)
		var cnt int
		cnt, err = reader.Read(bytes)
		if err == fio.EOF {
			err = nil
			res = stream.NewStepResultEmpty(stream.Empty[[]byte]())
		} else if err == nil {
			if cnt == 0 {
				res = stream.NewStepResultEmpty(stream.Empty[[]byte]())
			} else {
				res = stream.NewStepResult(bytes, ReadByteChunks(reader, chunkSize))
			}
		}
		return
	})
}

var emptyByteChunkStream = stream.Empty[[]byte]()

// SplitBySeparator splits byte-chunk stream by the given separator.
func SplitBySeparator(stm stream.Stream[[]byte], sep byte) stream.Stream[[]byte] {
	return stream.StateFlatMap(stm, []byte{}, func(a []byte, s []byte) (resultState []byte, stm stream.Stream[[]byte]) {
		parts := splitBy(sep, a, [][]byte{})
		if len(parts) == 0 {
			// stream finished??
			stm = emptyByteChunkStream
		} else if len(parts) == 1 {
			// separator not found
			resultState = append(s, a...)
			stm = emptyByteChunkStream
		} else {
			parts[0] = append(s, parts[0]...)
			resultState = parts[len(parts)-1]
			stm = stream.FromSlice(parts[0 : len(parts)-1])
		}
		return
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
func ReadLines(reader fio.Reader) stream.Stream[string] {
	chunks := ReadByteChunks(reader, DefaultChunkSize)
	rows := SplitBySeparator(chunks, '\n')
	return MapToStrings(rows)
}
