package gen_test

import (
	"compress/gzip"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4"
	"github.com/synapse-garden/phx/gen"
)

var _ gen.Compressor = new(gzip.Writer)
var _ gen.Compressor = new(snappy.Writer)
var _ gen.Compressor = new(lz4.Writer)
