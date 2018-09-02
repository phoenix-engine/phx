package cpp_test

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type mockCounter struct {
	int64
	io.WriteCloser
}

func (m mockCounter) Count() int64 { return m.int64 }

func makeLongBuffer() *bytes.Buffer {
	out := make([]byte, 80)
	for i := 0; i < 80; i++ {
		out[i] = byte(i)
	}
	return bytes.NewBuffer(out)
}

func makeRenderedLongBuffer() *bytes.Buffer {
	var (
		into = new(bytes.Buffer)
		in   = makeLongBuffer()
		bs   = in.Bytes()

		prev []byte
	)

	for i := 0; i < in.Len(); i += 11 {
		prev = nil
		for j := 0; j < 11 && j+i < in.Len(); j++ {
			prev = append(prev, bs[i+j])
			_, _ = into.WriteString(fmt.Sprintf(
				"0x%02x,", bs[i+j],
			))
		}
		_, _ = into.WriteString(" // |")
		_, _ = into.WriteString(replaceNonPrint(prev[:]))
		_, _ = into.WriteString("|\n")
	}

	return into
}

func replaceNonPrint(s []byte) string {
	var out []byte
	for _, c := range s {
		out = append(out, func() byte {
			if !unicode.IsPrint(rune(c)) ||
				unicode.IsSpace(rune(c)) {
				return '.'
			}
			return c
		}())
	}

	return string(out)
}
