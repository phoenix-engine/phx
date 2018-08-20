package gen_test

import "io"

var iw gen.ImplWriter
var _ = io.Writer(iw)
var _ = io.ReaderFrom(iw)

func Test() {}
