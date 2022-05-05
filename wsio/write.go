package wsio


import (
	fio "io"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

func WriteByteChunks(writer fio.ReadWriter, side ws.State, chunkSize int) stream.Sink[[]byte] {
	return func (stm stream.Stream[[]byte]) stream.Stream[io.Unit] { 
		return stream.MapEval(stm, func(data []byte) io.IO[io.Unit]{ 
			return io.Eval(func () (_ io.Unit, err error) {
				err = wsutil.WriteServerMessage(writer, ws.OpBinary, data)
				return 
			})
		})
	}
}
