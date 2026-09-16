package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Field is one row of a detail panel.
type Field struct {
	Label string
	Value string
}

// Meta is the API envelope, used when --meta is set.
type Meta struct {
	SchemaVersion string
	GeneratedAt   time.Time
	LastModified  *time.Time
	Total         *int
}

// View is the command-to-renderer seam.
type View struct {
	Data   any
	Meta   Meta
	Detail []Field
	Table  func() Table
}

// Options control how a View is printed.
type Options struct {
	Format Format
	Color  ColorMode
	Meta   bool
}

// Print writes v to w using the chosen format.
func Print(w io.Writer, opts Options, v View) error {
	if opts.Format.Tabular() {
		if opts.Format == FormatTable && len(v.Detail) > 0 {
			if err := renderDetail(w, v.Detail, usePretty(w, opts.Color)); err != nil {
				return err
			}
			if v.Table != nil {
				_, _ = fmt.Fprintln(w)
			}
		}
		var tbl Table
		if v.Table != nil {
			tbl = v.Table()
		}
		switch opts.Format {
		case FormatTable:
			return renderTable(w, opts.Color, tbl)
		case FormatCSV:
			return renderDelimited(w, tbl, ',')
		case FormatTSV:
			return renderDelimited(w, tbl, '\t')
		case FormatMarkdown:
			return renderMarkdown(w, tbl)
		}
	}

	payload := v.Data
	if opts.Meta {
		payload = wrapMeta(v)
	}
	switch opts.Format {
	case FormatJSON:
		return renderJSON(w, payload)
	case FormatYAML:
		return renderYAML(w, payload)
	default:
		return fmt.Errorf("unknown output format %q", opts.Format)
	}
}

func wrapMeta(v View) map[string]any {
	out := map[string]any{
		"schema_version": v.Meta.SchemaVersion,
		"generated_at":   v.Meta.GeneratedAt,
		"result":         v.Data,
	}
	if v.Meta.LastModified != nil {
		out["last_modified"] = v.Meta.LastModified
	}
	if v.Meta.Total != nil {
		out["total"] = v.Meta.Total
	}
	return out
}

func usePretty(w io.Writer, color ColorMode) bool {
	switch color {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		return isTTY(w)
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func renderDetail(w io.Writer, fields []Field, pretty bool) error {
	if len(fields) == 0 {
		return nil
	}
	width := 0
	for _, f := range fields {
		if n := len(f.Label); n > width {
			width = n
		}
	}
	r := lipgloss.NewRenderer(w)
	if pretty {
		r.SetColorProfile(termenv.TrueColor)
	} else {
		r.SetColorProfile(termenv.Ascii)
	}
	labelStyle := r.NewStyle().Faint(true)
	for _, f := range fields {
		label := f.Label + ":"
		if pretty {
			label = labelStyle.Render(fmt.Sprintf("%-*s", width+1, label))
		} else {
			label = fmt.Sprintf("%-*s", width+1, label)
		}
		if _, err := fmt.Fprintf(w, "%s  %s\n", label, f.Value); err != nil {
			return err
		}
	}
	return nil
}

func renderTable(w io.Writer, color ColorMode, t Table) error {
	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return nil
	}
	if usePretty(w, color) {
		return renderPrettyTable(w, t)
	}
	return renderPlainTable(w, t)
}

func renderPrettyTable(w io.Writer, t Table) error {
	r := lipgloss.NewRenderer(w)
	r.SetColorProfile(termenv.TrueColor)
	headerStyle := r.NewStyle().Bold(true).Padding(0, 1)
	base := r.NewStyle().Padding(0, 1)
	ok := r.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("2"))
	warn := r.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("3"))
	bad := r.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("1"))
	muted := r.NewStyle().Padding(0, 1).Faint(true)
	border := r.NewStyle().Foreground(lipgloss.Color("8"))

	headers := make([]string, len(t.Headers))
	for i, h := range t.Headers {
		headers[i] = strings.ToUpper(h)
	}
	rows := make([][]string, len(t.Rows))
	for i, row := range t.Rows {
		cells := make([]string, len(row))
		for j, c := range row {
			cells[j] = c.Text
		}
		rows[i] = cells
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(border).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row < 0 || row >= len(t.Rows) || col < 0 || col >= len(t.Rows[row]) {
				return base
			}
			switch t.Rows[row][col].Kind {
			case KindOK:
				return ok
			case KindWarn:
				return warn
			case KindBad:
				return bad
			case KindMuted:
				return muted
			default:
				return base
			}
		}).
		Headers(headers...).
		Rows(rows...)
	_, err := fmt.Fprintln(w, tbl.Render())
	return err
}

func renderPlainTable(w io.Writer, t Table) error {
	widths := make([]int, len(t.Headers))
	headers := make([]string, len(t.Headers))
	for i, h := range t.Headers {
		headers[i] = strings.ToUpper(h)
		widths[i] = lipgloss.Width(headers[i])
	}
	for _, row := range t.Rows {
		for i, c := range row {
			if i >= len(widths) {
				continue
			}
			if n := lipgloss.Width(c.Text); n > widths[i] {
				widths[i] = n
			}
		}
	}
	if err := writePlainRow(w, headers, widths, t.Aligns); err != nil {
		return err
	}
	for _, row := range t.Rows {
		cells := make([]string, len(row))
		for i, c := range row {
			cells[i] = c.Text
		}
		if err := writePlainRow(w, cells, widths, t.Aligns); err != nil {
			return err
		}
	}
	return nil
}

func writePlainRow(w io.Writer, cells []string, widths []int, aligns []Align) error {
	parts := make([]string, 0, len(cells))
	for i, c := range cells {
		width := 0
		if i < len(widths) {
			width = widths[i]
		}
		align := AlignLeft
		if i < len(aligns) {
			align = aligns[i]
		}
		parts = append(parts, pad(c, width, align))
	}
	_, err := fmt.Fprintln(w, strings.Join(parts, "  "))
	return err
}

func pad(s string, width int, align Align) string {
	n := lipgloss.Width(s)
	if n >= width {
		return s
	}
	gap := strings.Repeat(" ", width-n)
	if align == AlignRight {
		return gap + s
	}
	return s + gap
}

func renderDelimited(w io.Writer, t Table, comma rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = comma
	if err := cw.Write(t.Headers); err != nil {
		return err
	}
	for _, row := range t.Rows {
		rec := make([]string, len(row))
		for i, c := range row {
			rec[i] = c.Text
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func renderMarkdown(w io.Writer, t Table) error {
	if len(t.Headers) == 0 {
		return nil
	}
	write := func(cells []string) error {
		_, err := fmt.Fprintf(w, "| %s |\n", strings.Join(cells, " | "))
		return err
	}
	if err := write(t.Headers); err != nil {
		return err
	}
	sep := make([]string, len(t.Headers))
	for i := range sep {
		if i < len(t.Aligns) && t.Aligns[i] == AlignRight {
			sep[i] = "---:"
		} else {
			sep[i] = "---"
		}
	}
	if err := write(sep); err != nil {
		return err
	}
	for _, row := range t.Rows {
		cells := make([]string, len(t.Headers))
		for i := range cells {
			if i < len(row) {
				cells[i] = strings.ReplaceAll(row[i].Text, "|", "\\|")
			}
		}
		if err := write(cells); err != nil {
			return err
		}
	}
	return nil
}

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func renderYAML(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var generic any
	if err := json.Unmarshal(b, &generic); err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(generic); err != nil {
		return err
	}
	return enc.Close()
}
