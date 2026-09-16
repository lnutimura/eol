package cmd

import (
	"fmt"
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/spf13/cobra"
)

func (a *app) checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <product>[@<release>]",
		Short: "Check whether a release is past end-of-life",
		Long: `Check whether a product release cycle is end-of-life.

Exits 0 if the cycle is still in standard support, 2 if it is
end-of-life (isEol or unmaintained), and 1 on error.
Bare <product> checks the latest release cycle. An unknown @release
falls back to the longest matching cycle prefix, so ubuntu@22.04.1
resolves to 22.04.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeCheck,
		RunE: func(cmd *cobra.Command, args []string) error {
			product, release, _ := strings.Cut(args[0], "@")
			product = strings.TrimSpace(product)
			release = strings.TrimSpace(release)
			if product == "" {
				return fmt.Errorf("product name is required")
			}

			ctx, cancel := a.ctx(cmd)
			defer cancel()

			var rel eol.ProductRelease
			if release == "" {
				resp, err := a.client.GetLatestRelease(ctx, product)
				if err != nil {
					return err
				}
				rel = resp.Result
			} else {
				resp, err := a.client.GetProduct(ctx, product)
				if err != nil {
					return err
				}
				resolved, ok := resolveRelease(resp.Result.Releases, release)
				if !ok {
					return fmt.Errorf("release %q not found for product %q", release, product)
				}
				rel = resolved
			}

			status, code := checkStatus(rel)
			extra := ""
			if rel.EolFrom.Valid {
				extra = fmt.Sprintf(" (eol from %s)", rel.EolFrom.String())
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s is %s%s\n", product, rel.Name, status, extra)
			if code != 0 {
				return &ExitError{Code: code}
			}
			return nil
		},
	}
}

func checkStatus(rel eol.ProductRelease) (string, int) {
	if rel.IsEol || !rel.IsMaintained {
		return "end-of-life", 2
	}
	return "maintained", 0
}

func resolveRelease(releases []eol.ProductRelease, name string) (eol.ProductRelease, bool) {
	for _, r := range releases {
		if r.Name == name {
			return r, true
		}
	}
	var best eol.ProductRelease
	bestLen := -1
	for _, r := range releases {
		if !isReleasePrefix(name, r.Name) {
			continue
		}
		if len(r.Name) > bestLen {
			best = r
			bestLen = len(r.Name)
		}
	}
	if bestLen < 0 {
		return eol.ProductRelease{}, false
	}
	return best, true
}

func isReleasePrefix(query, cycle string) bool {
	if query == cycle {
		return true
	}
	if !strings.HasPrefix(query, cycle) {
		return false
	}
	return len(query) == len(cycle) || query[len(cycle)] == '.'
}
