package text_test

import (
	"bytes"
	fio "io"
	"io/fs"
	"os"
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
	"github.com/primetalk/goio/stream"
	"github.com/primetalk/goio/text"
	"github.com/stretchr/testify/assert"
)

const exampleText = `
Line 2
Line 30
`

func TestTextStream(t *testing.T) {
	data := []byte(exampleText)
	r := bytes.NewReader(data)
	strings := text.ReadLines(r)
	lens := stream.Map(strings, func(s string) int { return len(s) })
	lensSlice, err := io.UnsafeRunSync(stream.ToSlice(lens))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{0, 6, 7}, lensSlice)
	stream10_12 := stream.LiftMany(10, 11, 12)
	stream20_24 := stream.Map(stream10_12, func(i int) int { return i * 2 })
	res, err := io.UnsafeRunSync(stream.ToSlice(stream20_24))
	assert.NoError(t, err)
	assert.Equal(t, []int{20, 22, 24}, res)
}

func TestTextStreamWrite(t *testing.T) {
	data := []byte(exampleText)
	r := bytes.NewReader(data)
	strings := text.ReadLines(r)
	lens := stream.Map(strings, func(s string) int { return len(s) })
	lensAsString := stream.Map(lens, fun.ToString[int])
	w := bytes.NewBuffer([]byte{})
	writes := stream.ToSink(lensAsString, text.WriteLines(w))
	_, err := io.UnsafeRunSync(stream.DrainAll(writes))
	assert.NoError(t, err)
	assert.Equal(t, `0
6
7
`, w.String())
}

func TestFile(t *testing.T) {
	path := t.TempDir()+"/hello.txt"
	content := "hello"
	err := os.WriteFile(path, []byte(content), fs.ModePerm)
	assert.NoError(t, err)
	contentIO := resource.Use(text.ReadOnlyFile(path), func(f *os.File) io.IO[string] {
		return io.Eval(func() (str string, err error) {
			var bytes []byte
			bytes, err = fio.ReadAll(f)
			if err == nil {
				str = string(bytes)
			}
			return
		})
	})
	var str string
	str, err = io.UnsafeRunSync(contentIO)
	assert.NoError(t, err)
	assert.Equal(t, content, str)
}
