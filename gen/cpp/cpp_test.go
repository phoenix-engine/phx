package cpp_test

import (
	"io"

	"github.com/synapse-garden/phx/gen"
	"github.com/synapse-garden/phx/gen/cpp"
)

var (
	ct cpp.Target
	_  = gen.Encoder(ct)

	aw cpp.ArrayWriter
	_  = io.WriteCloser(aw)
	_  = io.ReaderFrom(aw)
)
