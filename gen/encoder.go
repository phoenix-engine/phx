package gen

import (
	"io"
)

type Encoder interface {
	Create(name string) (io.WriteCloser, error)
	Finalize() error
}
