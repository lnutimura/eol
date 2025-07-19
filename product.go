package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lutimura/eol/internal"
	"github.com/spf13/cobra"
)

type cmdProduct struct{}

func (c *cmdProduct) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "product"
	cmd.Short = "Query for all products"
	cmd.Long = "Query for all products referenced on endoflife.date"

	productGetCmd := cmdProductGet{}
	cmd.AddCommand(productGetCmd.Command())

	productListCmd := cmdProductList{}
	cmd.AddCommand(productListCmd.Command())

	// Workaround for subcommand usage errors. See: https://github.com/spf13/cobra/issues/706
	cmd.Args = cobra.NoArgs
	cmd.Run = func(cmd *cobra.Command, _ []string) { _ = cmd.Usage() }

	return cmd
}

type ProductListItem struct {
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases"`
	Label    string   `json:"label"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	URI      string   `json:"uri"`
}

type ProductListResponse struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   string            `json:"generated_at"`
	Total         int               `json:"total"`
	Result        []ProductListItem `json:"result"`
}
type cmdProductList struct {
	flagAllColumns bool
	flagColumns    []string

	flagCategory string
}

func (c *cmdProductList) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "list"
	cmd.Short = "List all products"
	cmd.Long = "List all the products referenced on endoflife.date"

	cmd.Flags().BoolVarP(&c.flagAllColumns, "all", "a", false, "Display all columns")
	cmd.Flags().StringSliceVarP(&c.flagColumns, "columns", "c", nil, "Comma-separated list of columns to display")

	cmd.Flags().StringVar(&c.flagCategory, "category", "", "Filter by category")

	cmd.Run = c.Run

	return cmd
}

func (c *cmdProductList) Run(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("%s/products", EndOfLifeURL)
	if c.flagCategory != "" {
		url = fmt.Sprintf("%s/categories/%s", EndOfLifeURL, c.flagCategory)
	}

	var response ProductListResponse
	if err := internal.FetchJSON(url, &response); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	columns := internal.ParseColumns[ProductListItem](
		c.flagAllColumns,
		c.flagColumns,
		[]string{"Name"},
	)
	internal.RenderTable(response.Result, columns)
}

type ProductGetResponse struct {
	SchemaVersion string    `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	LastModified  time.Time `json:"last_modified"`
	Result        Product   `json:"result"`
}

type Product struct {
	Name           string               `json:"name"`
	Aliases        []string             `json:"aliases"`
	Label          string               `json:"label"`
	Category       string               `json:"category"`
	Tags           []string             `json:"tags"`
	VersionCommand string               `json:"versionCommand"`
	Identifiers    []any                `json:"identifiers"`
	Labels         ProductSupportLabels `json:"labels"`
	Links          ProductLinks         `json:"links"`
	Releases       []ProductRelease     `json:"releases"`
}

type ProductSupportLabels struct {
	Eoas         string `json:"eoas"`
	Discontinued string `json:"discontinued"`
	Eol          string `json:"eol"`
	Eoes         string `json:"eoes"`
}

type ProductLinks struct {
	Icon          string `json:"icon"`
	HTML          string `json:"html"`
	ReleasePolicy string `json:"releasePolicy"`
}

type ProductRelease struct {
	Name         string               `json:"name"`
	Codename     string               `json:"codename"`
	Label        string               `json:"label"`
	ReleaseDate  string               `json:"releaseDate"`
	IsLts        bool                 `json:"isLts"`
	LtsFrom      string               `json:"ltsFrom"`
	IsEol        bool                 `json:"isEol"`
	EolFrom      string               `json:"eolFrom"`
	IsEoes       bool                 `json:"isEoes"`
	EoesFrom     string               `json:"eoesFrom"`
	IsMaintained bool                 `json:"isMaintained"`
	Latest       ProductLatestRelease `json:"latest"`
}

type ProductLatestRelease struct {
	Name string `json:"name"`
	Date string `json:"date"`
	Link string `json:"link"`
}

type cmdProductGet struct {
	flagAllColumns bool
	flagColumns    []string
}

func (c *cmdProductGet) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "get"
	cmd.Short = "Get a product"
	cmd.Long = "Get the given product data."
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().BoolVarP(&c.flagAllColumns, "all", "a", false, "Display all columns")
	cmd.Flags().StringSliceVarP(&c.flagColumns, "columns", "c", nil, "Comma-separated list of columns to display")

	cmd.Run = c.Run

	return cmd
}

func (c *cmdProductGet) Run(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("%s/products/%s", EndOfLifeURL, args[0])

	var response ProductGetResponse
	if err := internal.FetchJSON(url, &response); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	columns := internal.ParseColumns[ProductRelease](
		c.flagAllColumns,
		c.flagColumns,
		[]string{"Name", "Label", "ReleaseDate", "EolFrom", "EoesFrom"},
	)

	internal.RenderTable(response.Result.Releases, columns)
}
