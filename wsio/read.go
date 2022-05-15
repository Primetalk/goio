package wsio

import (
	fio "io"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

// FromWebSocket constructs a stream from ReadWriter
func FromWebSocket(r fio.ReadWriter, side ws.State, chunkSize int) stream.Stream[[]byte] {
	return io.Eval(func() (sr stream.StepResult[[]byte], err error) {
		var data []byte
		var opCode ws.OpCode
		data, opCode, err = wsutil.ReadData(r, side)
		if err == nil {
			isFinished := opCode == ws.OpClose
			var cont stream.Stream[[]byte]
			if isFinished {
				cont = stream.Empty[[]byte]()
			} else {
				cont = FromWebSocket(r, side, chunkSize)
			}
			sr = stream.NewStepResult(data, cont)
		}
		
		return	
	})
}
