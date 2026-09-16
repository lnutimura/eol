package cmd

import (
	"strings"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/spf13/cobra"
)

func (a *app) completeProducts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return a.completeProductNames(cmd, toComplete)
}

func (a *app) completeCategories(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return a.completeURINames(toComplete, func() ([]string, error) {
		resp, err := a.completionClient().ListCategories(cmd.Context())
		if err != nil {
			return nil, err
		}
		names := make([]string, len(resp.Result))
		for i, u := range resp.Result {
			names[i] = u.Name
		}
		return names, nil
	})
}

func (a *app) completeTags(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return a.completeURINames(toComplete, func() ([]string, error) {
		resp, err := a.completionClient().ListTags(cmd.Context())
		if err != nil {
			return nil, err
		}
		names := make([]string, len(resp.Result))
		for i, u := range resp.Result {
			names[i] = u.Name
		}
		return names, nil
	})
}

func (a *app) completeIdentifierTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return a.completeURINames(toComplete, func() ([]string, error) {
		resp, err := a.completionClient().ListIdentifierTypes(cmd.Context())
		if err != nil {
			return nil, err
		}
		names := make([]string, len(resp.Result))
		for i, u := range resp.Result {
			names[i] = u.Name
		}
		return names, nil
	})
}

func (a *app) completeCategoryFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return a.completeCategories(cmd, nil, toComplete)
}

func (a *app) completeTagFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return a.completeTags(cmd, nil, toComplete)
}

func (a *app) completeReleaseArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return a.completeProductNames(cmd, toComplete)
	case 1:
		return a.completeReleaseNames(cmd, args[0], toComplete)
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func (a *app) completeCheck(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	product, rest, found := strings.Cut(toComplete, "@")
	if !found {
		return a.completeProductNames(cmd, toComplete)
	}
	names, dir := a.completeReleaseNames(cmd, product, rest)
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = product + "@" + n
	}
	return out, dir
}

func (a *app) completeProductNames(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := a.completionClient().ListProducts(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, p := range resp.Result {
		if strings.HasPrefix(p.Name, toComplete) {
			if p.Label != "" && p.Label != p.Name {
				out = append(out, p.Name+"\t"+p.Label)
			} else {
				out = append(out, p.Name)
			}
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func (a *app) completeReleaseNames(cmd *cobra.Command, product, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := a.completionClient().GetProduct(cmd.Context(), product)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, r := range resp.Result.Releases {
		if strings.HasPrefix(r.Name, toComplete) {
			if r.Label != "" && r.Label != r.Name {
				out = append(out, r.Name+"\t"+r.Label)
			} else {
				out = append(out, r.Name)
			}
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func (a *app) completeURINames(toComplete string, list func() ([]string, error)) ([]string, cobra.ShellCompDirective) {
	names, err := list()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, n := range names {
		if strings.HasPrefix(n, toComplete) {
			out = append(out, n)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func (a *app) completionClient() *eol.Client {
	if a.client != nil {
		return a.client
	}
	a.client = a.newClient()
	return a.client
}
