package cpp_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"unicode"

	"github.com/synapse-garden/phx/gen/cpp"
	pt "github.com/synapse-garden/phx/testing"
)

var (
	ct cpp.Target
	_  = io.Writer(ct)
	_  = io.ReaderFrom(ct)
)

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

func TestArrayWriter(t *testing.T) {
	for i, test := range []struct {
		should    string
		given     io.Reader
		expect    []byte
		expectN   int64
		expectErr string
	}{{
		should: "encode an empty buffer as an empty literal",
		given:  bytes.NewBufferString(""),
		expect: []byte(""),
	}, {
		should:  "render a few bytes",
		given:   bytes.NewBufferString("hello"),
		expect:  []byte("0x68,0x65,0x6c,0x6c,0x6f, // |hello|\n"),
		expectN: 5,
	}, {
		should:  "render many bytes",
		given:   makeLongBuffer(),
		expect:  makeRenderedLongBuffer().Bytes(),
		expectN: 80,
	}} {
		t.Logf("test %d: should %s", i, test.should)

		var tmp bytes.Buffer
		aw := cpp.NewArrayWriter(&tmp)

		n, err := io.Copy(aw, test.given)
		if !pt.CheckErrMatches(t, err, test.expectErr) {
			continue
		}

		err = aw.Flush()

		if !pt.CheckErrMatches(t, err, test.expectErr) {
			continue
		}

		if !pt.CheckEq(t, n, test.expectN) {
			continue
		}

		if !pt.CheckBufEq(t, tmp.Bytes(), test.expect) {
			continue
		}
	}
}
