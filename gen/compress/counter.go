package compress

import "io"

type Counter interface{ Count() int64 }

type WCounter struct {
	Written int64
	io.Writer
}

func (c *WCounter) Count() int64 { return c.Written }

func (c *WCounter) Write(some []byte) (n int, err error) {
	n, err = c.Writer.Write(some)
	c.Written += int64(n)
	return
}

func (c *WCounter) ReadFrom(some io.Reader) (n int64, err error) {
	if rf, ok := c.Writer.(io.ReaderFrom); ok {
		n, err = rf.ReadFrom(some)
	} else {
		n, err = io.Copy(c.Writer, some)
	}

	c.Written += n
	return
}
