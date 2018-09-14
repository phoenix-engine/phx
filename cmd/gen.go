// Copyright © 2018 Bodie Solomon <bodie@synapsegarden.net>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"github.com/phoenix-engine/phx/fs"
	"github.com/phoenix-engine/phx/gen"
	"github.com/phoenix-engine/phx/gen/compress"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	from = new(string)
	to   = new(string)

	level = new(int)
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate build deps",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		pipeline := gen.Gen{
			From:  fs.Real{Where: *from},
			To:    fs.Real{Where: *to},
			Level: compress.Fastest,
			// func() gen.Level {
			// switch *level {
			// case 0:
			// 	return gen.Fastest
			// case 1:
			// 	return gen.Medium
			// case 2, 3:
			// 	return gen.High
			// case 9:
			// 	return gen.LZ4HC
			// default:
			// 	return gen.Medium
			// }
			// }(),
		}

		if err := pipeline.Operate(); err != nil {
			return errors.Wrap(err, "operating gen pipeline")
		}

		// TODO: Trigger dep update

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.PersistentFlags().StringVar(
		from, "from",
		"res",
		"Where to read static resources",
	)
	genCmd.PersistentFlags().StringVar(
		to, "to",
		"gen",
		"Where to write generated resources",
	)
	genCmd.PersistentFlags().IntVarP(
		level, "level", "l",
		0,
		"The compression level to use (0, 1, 2, 3, 9) "+
			"NOTE: > 0 unsupported due to bug",
	)
}
