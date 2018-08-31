package compress_test

var _ = compress.Compressor(new(compress.NoCompress))
var _ = compress.Maker(compress.NoMaker{})
