package cmd_test

import (
	"github.com/phoenix-engine/phx/cmd"
	"github.com/phoenix-engine/phx/path"

	"github.com/spf13/pflag"
)

var _ = path.Matcher(cmd.MatchAny{})
var _ = path.Matcher(cmd.Regexp{})
var _ = pflag.Value(new(cmd.Regexp))
