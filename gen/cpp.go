package cmd

import "io"

type ImplWriter struct{ Name string }

// Write implements io.Writer, encoding bytes into array literals.
func (i ImplWriter) Write(some []byte) (int, error) {
	return 0, nil
}

// ReadFrom implements io.ReaderFrom for the purpose of io.Copy.
func (i ImplWriter) ReadFrom(some io.Reader) (int64, error) {
	return 0, nil
}
