package compress_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/synapse-garden/phx/gen/compress"

	"github.com/pierrec/lz4"
)

func TestLZ4Maker(t *testing.T) {
	input := strings.Repeat("hello this is a rather long test ", 1000)

	inbuf := bytes.NewBufferString(input)
	comped := new(bytes.Buffer)

	madeLZ := compress.LZ4Maker{}.Make().(compress.Compressor)
	madeLZ.Reset(comped)

	n, err := io.Copy(madeLZ, inbuf)
	if err != nil {
		t.Errorf("Expected nil write error, but got %#v", err)
		t.FailNow()
	}

	if err := madeLZ.Close(); err != nil {
		t.Errorf("Expected nil error on Close, but got %#v", err)
		t.FailNow()
	}

	if exl := int64(len(input)); exl != n {
		t.Errorf("Expected %d bytes copied, but got %d", exl, n)
		t.FailNow()
	}

	if ct, ok := madeLZ.(compress.Counter); !ok {
		t.Errorf("Expected LZ4 to implement Counter, but got %T", madeLZ)
		t.FailNow()
	} else if num := ct.Count(); num != int64(comped.Len()) {
		t.Errorf("Expected count %d to equal consumed %d", num, n)
		t.FailNow()
	}

	zconsumer := lz4.NewReader(comped)

	outbuf := new(bytes.Buffer)

	n, err = io.Copy(outbuf, zconsumer)
	if err != nil {
		t.Errorf("Expected nil read error, but got %#v", err)
		t.FailNow()
	}

	if outbuf.String() != input {
		t.Error("Expected output to match input")
		t.FailNow()
	}
}
