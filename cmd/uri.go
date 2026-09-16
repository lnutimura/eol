package cmd

import (
	"context"
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
	"github.com/spf13/cobra"
)

func parentNoArgs(use, short, long string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	return cmd
}

type uriListSpec struct {
	use   string
	short string
	long  string
	fetch func(context.Context, *eol.Client) (eol.Response[[]eol.Uri], error)
}

func (a *app) uriListCmd(spec uriListSpec) *cobra.Command {
	return &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Long:  spec.long + "\n\nColumns: " + strings.Join(uriColumns.Names(), ", "),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := spec.fetch(ctx, a.client)
			if err != nil {
				return err
			}
			cols, err := uriColumns.Pick(a.all, a.columns, defaultURICols)
			if err != nil {
				return err
			}
			return a.print(cmd, output.View{
				Data: resp.Result,
				Meta: metaFrom(resp),
				Table: func() output.Table {
					return output.Project(resp.Result, cols)
				},
			})
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}
}

func (a *app) indexCmd() *cobra.Command {
	return a.uriListCmd(uriListSpec{
		use:   "index",
		short: "List the main API endpoints",
		long:  "List the main endoflife.date API endpoints.",
		fetch: func(ctx context.Context, c *eol.Client) (eol.Response[[]eol.Uri], error) {
			return c.Index(ctx)
		},
	})
}
