// Package cmd implements the eol command-line interface.
package cmd

import (
	"time"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/spf13/cobra"
)

type app struct {
	version string
	client  *eol.Client

	output   string
	color    string
	columns  []string
	all      bool
	meta     bool
	timeout  time.Duration
	cacheTTL time.Duration
	noCache  bool
	apiURL   string
}

// NewRootCmd constructs the eol command tree.
func NewRootCmd(version string) *cobra.Command {
	if version == "" {
		version = "dev"
	}
	a := &app{
		version:  version,
		output:   "table",
		color:    "auto",
		timeout:  eol.DefaultTimeout,
		cacheTTL: eol.DefaultTTL,
		apiURL:   eol.DefaultBaseURL,
	}

	cmd := &cobra.Command{
		Use:   "eol",
		Short: "A CLI for endoflife.date",
		Long: `eol queries the endoflife.date API for product support lifecycles.

Tabular formats (table, csv, tsv, markdown) honor --columns and --all.
JSON and YAML emit the API result payload; pass --meta to include
schema_version, generated_at, and last_modified.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			a.client = a.newClient()
			return nil
		},
	}
	cmd.SetVersionTemplate("eol {{.Version}}\n")

	f := cmd.PersistentFlags()
	f.StringVarP(&a.output, "output", "o", "table", "Output format: table, json, yaml, csv, tsv, markdown")
	f.StringVar(&a.color, "color", "auto", "When to color table output: auto, always, never")
	f.StringSliceVarP(&a.columns, "columns", "c", nil, "Comma-separated columns for tabular output")
	f.BoolVarP(&a.all, "all", "a", false, "Show all columns in tabular output")
	f.BoolVar(&a.meta, "meta", false, "Include API metadata in JSON/YAML output")
	f.DurationVar(&a.timeout, "timeout", eol.DefaultTimeout, "HTTP timeout")
	f.DurationVar(&a.cacheTTL, "cache-ttl", eol.DefaultTTL, "Serve cached API responses newer than this")
	f.BoolVar(&a.noCache, "no-cache", false, "Bypass the on-disk ETag cache")
	f.StringVar(&a.apiURL, "api-url", eol.DefaultBaseURL, "API base URL")
	_ = f.MarkHidden("api-url")

	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions(
		[]string{"table", "json", "yaml", "csv", "tsv", "markdown"},
		cobra.ShellCompDirectiveNoFileComp,
	))
	_ = cmd.RegisterFlagCompletionFunc("color", cobra.FixedCompletions(
		[]string{"auto", "always", "never"},
		cobra.ShellCompDirectiveNoFileComp,
	))

	cmd.AddCommand(
		a.productCmd(),
		a.releaseCmd(),
		a.categoryCmd(),
		a.tagCmd(),
		a.identifierCmd(),
		a.indexCmd(),
		a.checkCmd(),
	)
	return cmd
}

// Execute runs the root command.
func Execute(version string) error {
	return NewRootCmd(version).Execute()
}
