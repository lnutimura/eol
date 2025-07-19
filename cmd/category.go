package cmd

import (
	"fmt"
	"os"

	"github.com/lutimura/eol/internal"
	"github.com/spf13/cobra"
)

type cmdCategory struct{}

func (c *cmdCategory) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "category"
	cmd.Short = "Query for products by category"
	cmd.Long = "Query for products referenced on endoflife.date by category"

	categoryListCmd := cmdCategoryList{}
	cmd.AddCommand(categoryListCmd.Command())

	// Workaround for subcommand usage errors. See: https://github.com/spf13/cobra/issues/706
	cmd.Args = cobra.NoArgs
	cmd.Run = func(cmd *cobra.Command, _ []string) { _ = cmd.Usage() }

	return cmd
}

type CategoryListItem struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type CategoryListResponse struct {
	SchemaVersion string             `json:"schema_version"`
	GeneratedAt   string             `json:"generated_at"`
	Total         int                `json:"total"`
	Result        []CategoryListItem `json:"result"`
}

type cmdCategoryList struct {
	flagAllColumns bool
	flagColumns    []string

	flagName []string
}

func (c *cmdCategoryList) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "list"
	cmd.Short = "List all categories"
	cmd.Long = "List all endoflife.date categories"

	cmd.Flags().BoolVarP(&c.flagAllColumns, "all", "a", false, "Display all columns")
	cmd.Flags().StringSliceVarP(&c.flagColumns, "columns", "c", nil, "Comma-separated list of columns to display")

	cmd.Run = c.Run

	return cmd
}

func (c *cmdCategoryList) Run(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("%s/categories", internal.EndOfLifeURL)

	var response CategoryListResponse
	if err := internal.FetchJSON(url, &response); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	columns := internal.ParseColumns[CategoryListItem](
		c.flagAllColumns,
		c.flagColumns,
		[]string{"Name"},
	)
	internal.RenderTable(response.Result, columns)
}
