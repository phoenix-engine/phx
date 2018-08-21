package cpp_test

import (
	"bytes"
	"io"
	"testing"

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
	// each line == 72 + \n OR \r\n
	return new(bytes.Buffer)
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
