package text

import (
	"fmt"
	fio "io"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

// WriteByteChunks writes byte chunks to writer.
func WriteByteChunks(writer fio.Writer) stream.Sink[[]byte] {
	return func(stm stream.Stream[[]byte]) stream.Stream[fun.Unit] {
		return stream.MapEval(stm, func(data []byte) io.IO[fun.Unit] {
			return io.Eval(func() (_ fun.Unit, err error) {
				var cnt int
				cnt, err = writer.Write(data)
				if err == nil {
					if cnt != len(data) {
						err = fmt.Errorf("couldn't write %d bytes. Only %d was written", len(data), cnt)
					}
				}
				return
			})
		})
	}
}

// MapStringToBytes converts stream of strings to stream of byte chunks.
func MapStringToBytes(stm stream.Stream[string]) stream.Stream[[]byte] {
	return stream.Map(stm, func(s string) []byte { return []byte(s) })
}

var endline = []byte{'\n'}

// WriteLines creates a sink that receives strings and saves them to writer.
// It adds \n after each line.
func WriteLines(writer fio.Writer) stream.Sink[string] {
	return func(stm stream.Stream[string]) stream.Stream[fun.Unit] {
		bytes := MapStringToBytes(stm)
		bytesSep := stream.AddSeparatorAfterEachElement(bytes, endline)
		return WriteByteChunks(writer)(bytesSep)
	}
}
