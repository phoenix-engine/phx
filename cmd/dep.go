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

func init() {
	// TODO: Enable.
	// rootCmd.AddCommand(depCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// depCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// depCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
