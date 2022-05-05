package wsio

import (
	fio "io"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
)

func FromWebSocket(r fio.ReadWriter, side ws.State, chunkSize int) stream.Stream[[]byte] {
	return fromWebSocketImpl{
		r: r,
		chunkSize: chunkSize,
		side:  side,
	}
}

type fromWebSocketImpl struct {
	r fio.ReadWriter
	chunkSize int
	side ws.State
}

func (i fromWebSocketImpl)Step() (io.IO[stream.StepResult[[]byte]]) {
	return io.Eval(func() (sr stream.StepResult[[]byte], err error) {
		var data []byte
		var opCode ws.OpCode
		data, opCode, err = wsutil.ReadData(i.r, i.side)
		if err == nil {
			isFinished := opCode == ws.OpClose
			var cont stream.Stream[[]byte]
			if isFinished {
				cont = stream.Empty[[]byte]()
			} else {
				cont = i
			}
			sr = stream.NewStepResult(data, cont)
		}
		
		return	
	})
}

func (i fromWebSocketImpl)IsFinished() io.IO[bool] { 
	return io.Lift(false)// continuation will become empty when finished.
}
