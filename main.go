package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func createApp() *cobra.Command {
	app := &cobra.Command{}
	app.Use = "eol"
	app.Short = "eol is a cli for endoflife.date"
	app.Long = "eol is a cli for endoflife.date"

	// product sub-command
	productCmd := cmdProduct{}
	app.AddCommand(productCmd.Command())

	return app
}

func main() {
	app := createApp()
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while executing eol: %s\n", err)
		os.Exit(1)
	}
}
