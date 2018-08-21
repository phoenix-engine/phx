package gen

import (
	"io"
)

// Target defines the output formatting of a target language (e.g. C++).
type Target interface {
	io.ReaderFrom
	io.Reader
}
