package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var depCmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage lib dependencies",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("dep not implemented")
	},
}
