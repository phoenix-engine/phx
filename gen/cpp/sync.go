package cpp

import (
	"io"

	"github.com/synapse-garden/phx/gen/compress"
)

type DoneCloser struct {
	io.WriteCloser
	done chan<- struct{}
}

func (d DoneCloser) Close() error {
	defer close(d.done)
	return d.WriteCloser.Close()
}

func (d DoneCloser) Count() int64 {
	if ct, ok := d.WriteCloser.(compress.Counter); ok {
		return ct.Count()
	}
	return 0
}
