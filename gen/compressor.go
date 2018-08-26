package gen

import (
	"io"

	"github.com/pierrec/lz4"
)

// Compressor implements a simple common interface across compressors.
// Note that some compressors require a call to Close to finalize the
// stream.  They should have a wrapper type implemented in Flush.
type Compressor interface {
	Reset(to io.Writer)
	Write([]byte) (int, error)
	Flush() error
}

// NoCompress is a noop Compressor.
type NoCompress struct{ io.Writer }

// Reset implements Compressor on NoCompress.
func (n *NoCompress) Reset(to io.Writer) { n.Writer = to }

// Flush implements Compressor on NoCompress.
func (NoCompress) Flush() error { return nil }

// LZ4 is a wrapper for lz4.Writer which calls Close on Flush.
type LZ4 struct{ *lz4.Writer }

// Flush flushes and finalizes the LZ4 block stream.
func (l LZ4) Flush() error { return l.Writer.Close() }
