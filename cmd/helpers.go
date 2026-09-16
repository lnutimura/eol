package cmd

import (
	"context"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) ctx(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	timeout := a.timeout
	if timeout <= 0 {
		timeout = eol.DefaultTimeout
	}
	return context.WithTimeout(cmd.Context(), timeout)
}

func (a *app) newClient() *eol.Client {
	timeout := a.timeout
	if timeout <= 0 {
		timeout = eol.DefaultTimeout
	}
	return eol.New(eol.Options{
		BaseURL:      a.apiURL,
		UserAgent:    "eol/" + a.version,
		CacheTTL:     a.cacheTTL,
		DisableCache: a.noCache,
		Timeout:      timeout,
	})
}

func (a *app) print(cmd *cobra.Command, v output.View) error {
	format, err := output.ParseFormat(a.output)
	if err != nil {
		return err
	}
	color, err := output.ParseColor(a.color)
	if err != nil {
		return err
	}
	return output.Print(cmd.OutOrStdout(), output.Options{
		Format: format,
		Color:  color,
		Meta:   a.meta,
	}, v)
}

func metaFrom[T any](resp eol.Response[T]) output.Meta {
	return output.Meta{
		SchemaVersion: resp.SchemaVersion,
		GeneratedAt:   resp.GeneratedAt,
		LastModified:  resp.LastModified,
		Total:         resp.Total,
	}
}
