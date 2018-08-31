package compress_test

import (
	"compress/gzip"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4"
	"github.com/synapse-garden/phx/gen/compress"
)

var _ compress.Compressor = new(compress.NoCompress)
var _ compress.Compressor = new(gzip.Writer)
var _ compress.Compressor = new(snappy.Writer)
var _ compress.Compressor = new(lz4.Writer)

var _ = compress.Maker(compress.NoMaker{})
var _ = compress.Maker(compress.LZ4Maker{})
