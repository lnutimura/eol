package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eol",
	Short: "eol is a cli for endoflife.date",
	Long:  "eol is a cli for endoflife.date",
}

func Execute() error {
	productCmd := cmdProduct{}
	rootCmd.AddCommand(productCmd.Command())

	categoryCmd := cmdCategory{}
	rootCmd.AddCommand(categoryCmd.Command())
	return rootCmd.Execute()
}
