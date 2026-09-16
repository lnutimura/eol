// Package output renders CLI views as table, json, yaml, csv, tsv, or markdown.
package output

import (
	"fmt"
	"strings"
)

// Format is an output format.
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
	FormatCSV      Format = "csv"
	FormatTSV      Format = "tsv"
	FormatMarkdown Format = "markdown"
)

// ParseFormat parses a format name.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "csv":
		return FormatCSV, nil
	case "tsv":
		return FormatTSV, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unknown output format %q (valid: table, json, yaml, csv, tsv, markdown)", s)
	}
}

// Tabular reports whether the format uses columns rather than structured data.
func (f Format) Tabular() bool {
	switch f {
	case FormatTable, FormatCSV, FormatTSV, FormatMarkdown:
		return true
	default:
		return false
	}
}

// ColorMode controls ANSI color.
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// ParseColor parses a color mode.
func ParseColor(s string) (ColorMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return ColorAuto, nil
	case "always", "on", "true", "yes":
		return ColorAlways, nil
	case "never", "off", "false", "no":
		return ColorNever, nil
	default:
		return "", fmt.Errorf("unknown color mode %q (valid: auto, always, never)", s)
	}
}
