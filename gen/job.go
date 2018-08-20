package gen

import (
	"io"

	"github.com/synapse-garden/phx/fs"

	"github.com/golang/snappy"
	"github.com/pkg/errors"
)

type Job struct{ Name string }
type Done struct{ Name string }

func MakeChans() (chan Job, chan Done, chan struct{}, chan error) {
	return make(chan Job),
		make(chan Done),
		make(chan struct{}),
		make(chan error)
}

type Work struct {
	from, tmp fs.FS

	Jobs <-chan Job
	Done chan<- Done
	Kill <-chan struct{}
	Errs chan<- error

	Snap *snappy.Writer
}

func (w Work) Run() {
	for {
		select {
		case <-w.Kill:
			return

		case j, ok := <-w.Jobs:
			if !ok {
				return
			}

			done, err := w.Process(j.Name)
			if err != nil {
				select {
				case w.Errs <- err:
				case <-w.Kill:
				}
				return
			}

			select {
			case w.Done <- done:
			case <-w.Kill:
				return
			}
		}
	}
}

// Process encodes the file into a buffer using Snappy and returns it.
// When it is finished, the finished file is in w.tmp.
func (w Work) Process(path string) (Done, error) {
	var none Done

	ff, err := w.from.Open(path)
	if err != nil {
		return none, errors.Wrapf(err, "opening %s", path)
	}

	// Create the file which will contain the static resource class.
	buf, err := w.tmp.Create(path)
	if err != nil {
		return none, errors.Wrapf(err, "opening tempfile %s", path)
	}
	w.Snap.Reset(buf)

	if _, err := io.Copy(w.Snap, ff); err != nil {
		return none, errors.Wrapf(err, "encoding %s", path)
	}

	if err := w.Snap.Flush(); err != nil {
		return none, errors.Wrapf(err, "flushing %s", path)
	}

	if err := ff.Close(); err != nil {
		return none, errors.Wrapf(err, "closing input file %s", path)
	}

	if err := buf.Close(); err != nil {
		return none, errors.Wrapf(err, "closing tmpfile %s", path)
	}

	return Done{path}, nil
}
