package cmd

import (
	"context"
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) categoryCmd() *cobra.Command {
	cmd := parentNoArgs("category", "Query products by category", "Query products referenced on endoflife.date by category.")
	cmd.AddCommand(
		a.uriListCmd(uriListSpec{
			use:   "list",
			short: "List all categories",
			long:  "List all endoflife.date categories.",
			fetch: func(ctx context.Context, c *eol.Client) (eol.Response[[]eol.Uri], error) {
				return c.ListCategories(ctx)
			},
		}),
		a.categoryGetCmd(),
	)
	return cmd
}

func (a *app) categoryGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <category>",
		Short:             "List products in a category",
		Long:              "List all products in the given category.\n\nColumns: " + strings.Join(productSummaryColumns.Names(), ", "),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeCategories,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.ListProductsByCategory(ctx, args[0])
			if err != nil {
				return err
			}
			cols, err := productSummaryColumns.Pick(a.all, a.columns, defaultProductCols)
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
	}
}

func (a *app) tagCmd() *cobra.Command {
	cmd := parentNoArgs("tag", "Query products by tag", "Query products referenced on endoflife.date by tag.")
	cmd.AddCommand(
		a.uriListCmd(uriListSpec{
			use:   "list",
			short: "List all tags",
			long:  "List all endoflife.date tags.",
			fetch: func(ctx context.Context, c *eol.Client) (eol.Response[[]eol.Uri], error) {
				return c.ListTags(ctx)
			},
		}),
		a.tagGetCmd(),
	)
	return cmd
}

func (a *app) tagGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <tag>",
		Short:             "List products with a tag",
		Long:              "List all products with the given tag.\n\nColumns: " + strings.Join(productSummaryColumns.Names(), ", "),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeTags,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.ListProductsByTag(ctx, args[0])
			if err != nil {
				return err
			}
			cols, err := productSummaryColumns.Pick(a.all, a.columns, defaultProductCols)
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
	}
}

func (a *app) identifierCmd() *cobra.Command {
	cmd := parentNoArgs("identifier", "Query products by identifier", "Query products referenced on endoflife.date by identifier type.")
	cmd.AddCommand(
		a.uriListCmd(uriListSpec{
			use:   "list",
			short: "List identifier types",
			long:  "List all identifier types, such as purl, known in endoflife.date.",
			fetch: func(ctx context.Context, c *eol.Client) (eol.Response[[]eol.Uri], error) {
				return c.ListIdentifierTypes(ctx)
			},
		}),
		a.identifierGetCmd(),
	)
	return cmd
}

func (a *app) identifierGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <type>",
		Short:             "List identifiers of a given type",
		Long:              "List all identifiers of the given type, each with its product.\n\nColumns: " + strings.Join(identifierColumns.Names(), ", "),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeIdentifierTypes,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			resp, err := a.client.ListIdentifiers(ctx, args[0])
			if err != nil {
				return err
			}
			cols, err := identifierColumns.Pick(a.all, a.columns, defaultIdentCols)
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
	}
}
