package cmd

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lnutimura/eol/internal/eol"
	"github.com/lnutimura/eol/internal/output"
)

const eolWarnWindow = 120 * 24 * time.Hour

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func dashPtr(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}

func formatDate(d eol.Date) string {
	if !d.Valid {
		return "-"
	}
	return d.String()
}

func formatBool(b bool) string {
	return strconv.FormatBool(b)
}

func formatBoolPtr(b *bool) string {
	if b == nil {
		return "-"
	}
	return strconv.FormatBool(*b)
}

func joinOrDash(ss []string) string {
	if len(ss) == 0 {
		return "-"
	}
	return strings.Join(ss, ", ")
}

func mutedDash(s string) output.Kind {
	if s == "-" {
		return output.KindMuted
	}
	return output.KindPlain
}

func dateKind(d eol.Date) output.Kind {
	if !d.Valid {
		return output.KindMuted
	}
	return output.KindPlain
}

func releaseStatusKind(r eol.ProductRelease) output.Kind {
	if !r.IsMaintained || r.IsEol {
		return output.KindBad
	}
	if r.EolFrom.Valid {
		until := time.Until(r.EolFrom.Time)
		if until <= 0 {
			return output.KindBad
		}
		if until <= eolWarnWindow {
			return output.KindWarn
		}
	}
	return output.KindOK
}

func eolDateKind(r eol.ProductRelease) output.Kind {
	if !r.EolFrom.Valid {
		return output.KindMuted
	}
	return releaseStatusKind(r)
}

func boolKind(v bool, goodWhenTrue bool) output.Kind {
	if v == goodWhenTrue {
		return output.KindOK
	}
	return output.KindBad
}

func boolPtrKind(v *bool, goodWhenTrue bool) output.Kind {
	if v == nil {
		return output.KindMuted
	}
	return boolKind(*v, goodWhenTrue)
}

func latestName(r eol.ProductRelease) string {
	if r.Latest == nil {
		return "-"
	}
	return dash(r.Latest.Name)
}

func latestDate(r eol.ProductRelease) string {
	if r.Latest == nil {
		return "-"
	}
	return formatDate(r.Latest.Date)
}

func latestLink(r eol.ProductRelease) string {
	if r.Latest == nil {
		return "-"
	}
	return dashPtr(r.Latest.Link)
}

func formatCustom(m map[string]*string) string {
	if len(m) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if v := m[k]; v == nil {
			parts = append(parts, k+"=-")
		} else {
			parts = append(parts, k+"="+*v)
		}
	}
	return strings.Join(parts, ",")
}

func formatIdentifiers(ids []eol.Identifier) string {
	if len(ids) == 0 {
		return "-"
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.Type + ": " + id.ID
	}
	return strings.Join(parts, ", ")
}

var uriColumns = output.Columns[eol.Uri]{
	{Name: "name", Header: "Name", Value: func(u eol.Uri) string { return u.Name }},
	{Name: "uri", Header: "URI", Value: func(u eol.Uri) string { return u.URI }},
}

var productSummaryColumns = output.Columns[eol.ProductSummary]{
	{Name: "name", Header: "Name", Value: func(p eol.ProductSummary) string { return p.Name }},
	{Name: "aliases", Header: "Aliases", Value: func(p eol.ProductSummary) string { return joinOrDash(p.Aliases) }, Kind: func(p eol.ProductSummary) output.Kind { return mutedDash(joinOrDash(p.Aliases)) }},
	{Name: "label", Header: "Label", Value: func(p eol.ProductSummary) string { return p.Label }},
	{Name: "category", Header: "Category", Value: func(p eol.ProductSummary) string { return p.Category }},
	{Name: "tags", Header: "Tags", Value: func(p eol.ProductSummary) string { return joinOrDash(p.Tags) }},
	{Name: "uri", Header: "URI", Value: func(p eol.ProductSummary) string { return p.URI }},
}

var productFullColumns = output.Columns[eol.ProductDetails]{
	{Name: "name", Header: "Name", Value: func(p eol.ProductDetails) string { return p.Name }},
	{Name: "aliases", Header: "Aliases", Value: func(p eol.ProductDetails) string { return joinOrDash(p.Aliases) }, Kind: func(p eol.ProductDetails) output.Kind { return mutedDash(joinOrDash(p.Aliases)) }},
	{Name: "label", Header: "Label", Value: func(p eol.ProductDetails) string { return p.Label }},
	{Name: "category", Header: "Category", Value: func(p eol.ProductDetails) string { return p.Category }},
	{Name: "tags", Header: "Tags", Value: func(p eol.ProductDetails) string { return joinOrDash(p.Tags) }},
	{Name: "latest", Header: "Latest", Value: func(p eol.ProductDetails) string {
		if c := p.LatestCycle(); c != nil {
			return latestName(*c)
		}
		return "-"
	}},
	{Name: "latestDate", Header: "Latest Date", Value: func(p eol.ProductDetails) string {
		if c := p.LatestCycle(); c != nil {
			return latestDate(*c)
		}
		return "-"
	}},
	{Name: "releases", Header: "Releases", Value: func(p eol.ProductDetails) string {
		return strconv.Itoa(len(p.Releases))
	}, Align: output.AlignRight},
}

var productReleaseColumns = output.Columns[eol.ProductRelease]{
	{Name: "name", Header: "Name", Value: func(r eol.ProductRelease) string { return r.Name }, Kind: releaseStatusKind},
	{Name: "codename", Header: "Codename", Value: func(r eol.ProductRelease) string { return dashPtr(r.Codename) }, Kind: func(r eol.ProductRelease) output.Kind { return mutedDash(dashPtr(r.Codename)) }},
	{Name: "label", Header: "Label", Value: func(r eol.ProductRelease) string { return r.Label }},
	{Name: "releaseDate", Header: "Release Date", Value: func(r eol.ProductRelease) string { return formatDate(r.ReleaseDate) }, Kind: func(r eol.ProductRelease) output.Kind { return dateKind(r.ReleaseDate) }},
	{Name: "isLts", Header: "LTS", Value: func(r eol.ProductRelease) string { return formatBool(r.IsLts) }},
	{Name: "ltsFrom", Header: "LTS From", Value: func(r eol.ProductRelease) string { return formatDate(r.LtsFrom) }, Kind: func(r eol.ProductRelease) output.Kind { return dateKind(r.LtsFrom) }},
	{Name: "isEoas", Header: "EOAS", Value: func(r eol.ProductRelease) string { return formatBoolPtr(r.IsEoas) }, Kind: func(r eol.ProductRelease) output.Kind { return boolPtrKind(r.IsEoas, false) }},
	{Name: "eoasFrom", Header: "EOAS From", Value: func(r eol.ProductRelease) string { return formatDate(r.EoasFrom) }, Kind: func(r eol.ProductRelease) output.Kind { return dateKind(r.EoasFrom) }},
	{Name: "isEol", Header: "EOL", Value: func(r eol.ProductRelease) string { return formatBool(r.IsEol) }, Kind: func(r eol.ProductRelease) output.Kind { return boolKind(r.IsEol, false) }},
	{Name: "eolFrom", Header: "EOL From", Value: func(r eol.ProductRelease) string { return formatDate(r.EolFrom) }, Kind: eolDateKind},
	{Name: "isDiscontinued", Header: "Discontinued", Value: func(r eol.ProductRelease) string { return formatBoolPtr(r.IsDiscontinued) }, Kind: func(r eol.ProductRelease) output.Kind { return boolPtrKind(r.IsDiscontinued, false) }},
	{Name: "discontinuedFrom", Header: "Discontinued From", Value: func(r eol.ProductRelease) string { return formatDate(r.DiscontinuedFrom) }, Kind: func(r eol.ProductRelease) output.Kind { return dateKind(r.DiscontinuedFrom) }},
	{Name: "isEoes", Header: "EOES", Value: func(r eol.ProductRelease) string { return formatBoolPtr(r.IsEoes) }, Kind: func(r eol.ProductRelease) output.Kind { return boolPtrKind(r.IsEoes, false) }},
	{Name: "eoesFrom", Header: "EOES From", Value: func(r eol.ProductRelease) string { return formatDate(r.EoesFrom) }, Kind: func(r eol.ProductRelease) output.Kind { return dateKind(r.EoesFrom) }},
	{Name: "isMaintained", Header: "Maintained", Value: func(r eol.ProductRelease) string { return formatBool(r.IsMaintained) }, Kind: func(r eol.ProductRelease) output.Kind { return boolKind(r.IsMaintained, true) }},
	{Name: "latest", Header: "Latest", Value: latestName, Kind: func(r eol.ProductRelease) output.Kind { return mutedDash(latestName(r)) }},
	{Name: "latestDate", Header: "Latest Date", Value: latestDate, Kind: func(r eol.ProductRelease) output.Kind { return mutedDash(latestDate(r)) }},
	{Name: "latestLink", Header: "Latest Link", Value: latestLink, Kind: func(r eol.ProductRelease) output.Kind { return mutedDash(latestLink(r)) }},
	{Name: "custom", Header: "Custom", Value: func(r eol.ProductRelease) string { return formatCustom(r.Custom) }, Kind: func(r eol.ProductRelease) output.Kind { return mutedDash(formatCustom(r.Custom)) }},
}

var identifierColumns = output.Columns[eol.IdentifierRef]{
	{Name: "identifier", Header: "Identifier", Value: func(r eol.IdentifierRef) string { return r.Identifier }},
	{Name: "product", Header: "Product", Value: func(r eol.IdentifierRef) string { return r.Product.Name }},
	{Name: "uri", Header: "URI", Value: func(r eol.IdentifierRef) string { return r.Product.URI }},
}

var (
	defaultURICols     = []string{"name"}
	defaultProductCols = []string{"name"}
	defaultFullCols    = []string{"name", "latest", "latestDate", "releases"}
	defaultReleaseCols = []string{"name", "label", "releaseDate", "eolFrom", "eoesFrom"}
	defaultIdentCols   = []string{"identifier", "product"}
)
