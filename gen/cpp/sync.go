package cpp

import "io"

type DoneCloser struct {
	io.WriteCloser
	done chan<- struct{}
}

func (d DoneCloser) Close() error {
	defer close(d.done)
	return d.WriteCloser.Close()
}
