package cpp_test

import (
	"io"

	"github.com/phoenix-engine/phx/gen"
	"github.com/phoenix-engine/phx/gen/cpp"
)

var (
	ct cpp.Target
	_  = gen.Encoder(ct)

	aw cpp.ArrayWriter
	_  = io.WriteCloser(aw)
	_  = io.ReaderFrom(aw)
)
