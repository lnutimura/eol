package cmd

import (
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) productCmd() *cobra.Command {
	cmd := parentNoArgs("product", "Query products", "Query products referenced on endoflife.date.")
	cmd.AddCommand(a.productListCmd(), a.productGetCmd())
	return cmd
}

func (a *app) productListCmd() *cobra.Command {
	var category, tag string
	var full bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all products",
		Long: `List all products referenced on endoflife.date.

Columns: ` + strings.Join(productSummaryColumns.Names(), ", ") + `
With --full: ` + strings.Join(productFullColumns.Names(), ", "),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := a.ctx(cmd)
			defer cancel()
			if full {
				resp, err := a.client.ListProductsFull(ctx)
				if err != nil {
					return err
				}
				items := make([]eol.ProductDetails, 0, len(resp.Result))
				for _, p := range resp.Result {
					if matchCategory(p.Category, category) && matchTags(p.Tags, tag) {
						items = append(items, p)
					}
				}
				cols, err := productFullColumns.Pick(a.all, a.columns, defaultFullCols)
				if err != nil {
					return err
				}
				return a.print(cmd, output.View{
					Data: items,
					Meta: metaFrom(resp),
					Table: func() output.Table {
						return output.Project(items, cols)
					},
				})
			}

			var (
				resp eol.Response[[]eol.ProductSummary]
				err  error
			)
			switch {
			case category != "" && tag != "":
				resp, err = a.client.ListProductsByCategory(ctx, category)
				if err != nil {
					return err
				}
				filtered := make([]eol.ProductSummary, 0, len(resp.Result))
				for _, p := range resp.Result {
					if matchTags(p.Tags, tag) {
						filtered = append(filtered, p)
					}
				}
				resp.Result = filtered
			case category != "":
				resp, err = a.client.ListProductsByCategory(ctx, category)
			case tag != "":
				resp, err = a.client.ListProductsByTag(ctx, tag)
			default:
				resp, err = a.client.ListProducts(ctx)
			}
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
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().BoolVar(&full, "full", false, "Fetch full product details (unlocks latest, latestDate, releases)")
	_ = cmd.RegisterFlagCompletionFunc("category", a.completeCategoryFlag)
	_ = cmd.RegisterFlagCompletionFunc("tag", a.completeTagFlag)
	return cmd
}

func (a *app) productGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <product>",
		Short: "Get a product",
		Long: `Get the given product, including labels, links, identifiers, and releases.

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
			p := resp.Result
			return a.print(cmd, output.View{
				Data:   p,
				Meta:   metaFrom(resp),
				Detail: productDetail(p),
				Table: func() output.Table {
					return output.Project(p.Releases, cols)
				},
			})
		},
	}
}

func productDetail(p eol.ProductDetails) []output.Field {
	return []output.Field{
		{Label: "Name", Value: p.Name},
		{Label: "Label", Value: p.Label},
		{Label: "Category", Value: dash(p.Category)},
		{Label: "Tags", Value: joinOrDash(p.Tags)},
		{Label: "Aliases", Value: joinOrDash(p.Aliases)},
		{Label: "Version", Value: dashPtr(p.VersionCommand)},
		{Label: "Identifiers", Value: formatIdentifiers(p.Identifiers)},
		{Label: "EOAS label", Value: dashPtr(p.Labels.Eoas)},
		{Label: "Discontinued label", Value: dashPtr(p.Labels.Discontinued)},
		{Label: "EOL label", Value: dash(p.Labels.Eol)},
		{Label: "EOES label", Value: dashPtr(p.Labels.Eoes)},
		{Label: "HTML", Value: dash(p.Links.HTML)},
		{Label: "Icon", Value: dashPtr(p.Links.Icon)},
		{Label: "Release policy", Value: dashPtr(p.Links.ReleasePolicy)},
	}
}

func matchCategory(got, want string) bool {
	return want == "" || strings.EqualFold(got, want)
}

func matchTags(tags []string, want string) bool {
	if want == "" {
		return true
	}
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}
