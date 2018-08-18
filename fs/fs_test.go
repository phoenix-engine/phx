package fs_test

import (
	"github.com/synapse-garden/phx/fs"
)

var _ = fs.FS(fs.Real{})
var _ = fs.FS(fs.Mem{})
