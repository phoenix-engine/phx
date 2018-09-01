package gen

import (
	"io"

	"github.com/synapse-garden/phx/fs"
	"github.com/synapse-garden/phx/gen/compress"

	"github.com/pkg/errors"
)

type Job struct{ Name string }
type Done struct {
	Name                 string
	Size, CompressedSize int64
}

func MakeChans() (chan Job, chan Done, chan struct{}, chan error) {
	return make(chan Job),
		make(chan Done),
		make(chan struct{}),
		make(chan error)
}

type Work struct {
	from fs.FS

	Jobs <-chan Job
	Done chan<- Done
	Kill <-chan struct{}
	Errs chan<- error

	// The Encoder is responsible for creating and finalizing the
	// output files for the specific implementation.
	Encoder
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

// Process encodes the file into a buffer using LZ4 and returns it.
// When it is finished, the finished file is in w.tmp.
func (w Work) Process(path string) (none Done, err error) {
	ff, err := w.from.Open(path)
	if err != nil {
		return none, errors.Wrapf(err, "opening %s", path)
	}

	out, err := w.Encoder.Create(path)
	if err != nil {
		return none, errors.Wrapf(err, "opening tempfile %s", path)
	}

	// Encode the asset file using the Encoder's provided writer.
	n, err := io.Copy(out, ff)
	if err != nil {
		return none, errors.Wrapf(err, "encoding %s", path)
	}

	// Close the input file.
	if err := ff.Close(); err != nil {
		return none, errors.Wrapf(err, "closing input file %s", path)
	}

	// Close / Flush the Encoder's writer.  The implementation may
	// create or write more files, etc.
	if err := out.Close(); err != nil {
		return none, errors.Wrapf(err, "flushing compressor from %s", path)
	}

	done := Done{Name: path, Size: n}

	// Check for a compression counter.
	if c, ok := out.(compress.Counter); ok {
		done.CompressedSize = c.Count()
	}

	return done, nil
}
