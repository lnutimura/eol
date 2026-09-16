package cmd

import (
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) releaseCmd() *cobra.Command {
	cmd := parentNoArgs("release", "Query product releases", "Query product release cycles on endoflife.date.")
	cmd.AddCommand(a.releaseListCmd(), a.releaseGetCmd(), a.releaseLatestCmd())
	return cmd
}

func (a *app) releaseListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <product>",
		Short: "List release cycles for a product",
		Long: `List all release cycles for a product.

Columns: ` + strings.Join(productReleaseColumns.Names(), ", "),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeProducts,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.GetProduct(ctx, args[0])
			if err != nil {
				return err
			}
			cols, err := productReleaseColumns.Pick(a.all, a.columns, defaultReleaseCols)
			if err != nil {
				return err
			}
			return a.print(cmd, output.View{
				Data: resp.Result.Releases,
				Meta: metaFrom(resp),
				Table: func() output.Table {
					return output.Project(resp.Result.Releases, cols)
				},
			})
		},
	}
}

func (a *app) releaseGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <product> <release>",
		Short: "Get a product release cycle",
		Long: `Get a single product release cycle.

Columns: ` + strings.Join(productReleaseColumns.Names(), ", "),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: a.completeReleaseArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.GetRelease(ctx, args[0], args[1])
			if err != nil {
				return err
			}
			return a.printRelease(cmd, resp.Result, metaFrom(resp))
		},
	}
}

func (a *app) releaseLatestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "latest <product>",
		Short: "Get a product's latest release cycle",
		Long: `Get the latest release cycle for a product.

Columns: ` + strings.Join(productReleaseColumns.Names(), ", "),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeProducts,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.GetLatestRelease(ctx, args[0])
			if err != nil {
				return err
			}
			return a.printRelease(cmd, resp.Result, metaFrom(resp))
		},
	}
}

func (a *app) printRelease(cmd *cobra.Command, r eol.ProductRelease, meta output.Meta) error {
	cols, err := productReleaseColumns.Pick(a.all, a.columns, defaultReleaseCols)
	if err != nil {
		return err
	}
	return a.print(cmd, output.View{
		Data:   r,
		Meta:   meta,
		Detail: releaseDetail(r),
		Table: func() output.Table {
			return output.Project([]eol.ProductRelease{r}, cols)
		},
	})
}

func releaseDetail(r eol.ProductRelease) []output.Field {
	return []output.Field{
		{Label: "Name", Value: r.Name},
		{Label: "Label", Value: r.Label},
		{Label: "Codename", Value: dashPtr(r.Codename)},
		{Label: "Maintained", Value: formatBool(r.IsMaintained)},
		{Label: "LTS", Value: formatBool(r.IsLts)},
		{Label: "Release date", Value: formatDate(r.ReleaseDate)},
		{Label: "EOAS from", Value: formatDate(r.EoasFrom)},
		{Label: "EOL from", Value: formatDate(r.EolFrom)},
		{Label: "EOES from", Value: formatDate(r.EoesFrom)},
		{Label: "Latest", Value: latestName(r)},
		{Label: "Latest date", Value: latestDate(r)},
		{Label: "Latest link", Value: latestLink(r)},
	}
}
