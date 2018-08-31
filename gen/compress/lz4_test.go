package compress_test

var _ = compress.Maker(compress.LZ4Maker{})
var _ = compress.Compressor(compress.LZ4{})
