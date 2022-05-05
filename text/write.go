package text

import (
	"fmt"
	fio "io"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

func WriteByteChunks(writer fio.Writer, chunkSize int) stream.Sink[[]byte] {
	return func (stm stream.Stream[[]byte]) stream.Stream[io.Unit] { 
		return stream.MapEval(stm, func(data []byte) io.IO[io.Unit]{ 
			return io.Eval(func () (_ io.Unit, err error) {
				var cnt int
				cnt, err = writer.Write(data)
				if err == nil {
					if cnt != len(data) {
						err = fmt.Errorf("Couldn't write %d bytes. Only %d was written", len(data), cnt)
					}
				}
				return 
			})
		})
	}
}

func MapStringToBytes(stm stream.Stream[string])stream.Stream[[]byte] {
	return stream.Map(stm, func(s string)[]byte { return []byte(s) })
}

var endline = []byte{'\n'}

func WriteLines(writer fio.Writer) stream.Sink[string] {
	return func (stm stream.Stream[string]) stream.Stream[io.Unit] {
		bytes := MapStringToBytes(stm)
		bytesSep := stream.AddSeparatorAfterEachElement(bytes, endline)
		return WriteByteChunks(writer, DefaultChunkSize)(bytesSep)
	}
}
