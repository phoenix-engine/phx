package fs_test

import (
	"github.com/phoenix-engine/phx/fs"
)

var _ = fs.FS(fs.Real{})
var _ = fs.FS(fs.Mem{})
