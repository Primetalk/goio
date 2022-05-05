package text_test

import (
	"bytes"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/primetalk/goio/text"
	"github.com/stretchr/testify/assert"
)

const exampleText = `
Line 2
Line 3
`
func TestStream(t *testing.T) {
	data := []byte(exampleText)
	r := bytes.NewReader(data)
	strings := text.ReadLines(r)
	stream.Map strings
	empty := stream.Empty[int]()
	_, err := io.UnsafeRunSync(stream.DrainAll(empty))
	assert.Equal(t, err, nil)
	stream10_12 := stream.LiftMany(10, 11, 12)
	stream20_24 := stream.Map(stream10_12, func(i int)int { return i * 2 })
	res, err := io.UnsafeRunSync(stream.ToSlice(stream20_24))
	assert.Equal(t, err, nil)
	assert.Equal(t, res, []int{20,22,24})
}
