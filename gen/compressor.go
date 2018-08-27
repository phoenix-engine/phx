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

// LZ4 is a wrapper for lz4.Writer which calls Close on Flush.
type LZ4 struct {
	*lz4.Writer
	Level
}

// Flush flushes and finalizes the LZ4 block stream.
func (l LZ4) Flush() error { return l.Writer.Close() }
func (l LZ4) Reset(w io.Writer) {
	l.Writer.Reset(w)
	l.Writer.Header = lz4.Header{
		CompressionLevel: func() int {
			switch l.Level {
			case Fastest:
				return 0
			case Medium:
				return 3
			case LZ4HC, High:
				return 7
			default:
				return 3
			}
		}(),
	}
}
