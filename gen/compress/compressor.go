package compress

import (
	"io"
)

type Maker interface {
	Make() Compressor
}

type NoMaker struct{}

func (NoMaker) Make() Compressor {
	return &NoCompress{nil}
}

// Compressor implements a simple common interface across compressors.
// Note that some compressors require a call to Close to finalize the
// stream.  They should have a wrapper type implemented in Flush.
type Compressor interface {
	Reset(to io.Writer)
	Write([]byte) (int, error)
	Flush() error
	Close() error
}

// NoCompress is a noop Compressor.
type NoCompress struct{ io.Writer }

// Reset implements Compressor on NoCompress.
func (n *NoCompress) Reset(to io.Writer) { n.Writer = to }

func (NoCompress) Flush() error { return nil }

// Flush implements Compressor on NoCompress.
func (NoCompress) Close() error { return nil }

// Level is a mapping for compression levels.  Each Compressor may have
// its own way of mapping Level constants to internal values.  User-
// facing code should select a Level from among the defined constants.
type Level int

// Level constants.
const (
	Fastest Level = iota
	Medium
	High

	// LZ4-specific.
	LZ4HC
)

// Implement fmt.Stringer for fmt-compatible output strings.
func (l Level) String() string {
	switch l {
	case Fastest:
		return "fastest"
	case Medium:
		return "medium"
	case High:
		return "high"
	case LZ4HC:
		return "lz4hc"
	default:
		return "unknown"
	}
}
