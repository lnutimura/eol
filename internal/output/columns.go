package output

import (
	"fmt"
	"strings"
)

// Align is a column alignment.
type Align int

const (
	AlignLeft Align = iota
	AlignRight
)

// Kind is a display hint for a cell.
type Kind int

const (
	KindPlain Kind = iota
	KindMuted
	KindOK
	KindWarn
	KindBad
)

// Column describes one projected field.
type Column[T any] struct {
	Name   string
	Header string
	Value  func(T) string
	Align  Align
	Kind   func(T) Kind
}

// Columns is an ordered registry of columns for T.
type Columns[T any] []Column[T]

// Names returns the registry names in order.
func (cs Columns[T]) Names() []string {
	names := make([]string, len(cs))
	for i, c := range cs {
		names[i] = c.Name
	}
	return names
}

// Select returns the named columns in the given order.
// Matching is case-insensitive. Unknown names are an error.
func (cs Columns[T]) Select(names []string) (Columns[T], error) {
	if len(names) == 0 {
		return cs, nil
	}
	out := make(Columns[T], 0, len(names))
	var unknown []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		found := false
		for _, c := range cs {
			if strings.EqualFold(c.Name, name) {
				out = append(out, c)
				found = true
				break
			}
		}
		if !found {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		quoted := make([]string, len(unknown))
		for i, u := range unknown {
			quoted[i] = fmt.Sprintf("%q", u)
		}
		return nil, fmt.Errorf("unknown column %s (valid columns: %s)", strings.Join(quoted, ", "), strings.Join(cs.Names(), ", "))
	}
	return out, nil
}

// Pick chooses all columns, a named subset, or the defaults.
func (cs Columns[T]) Pick(all bool, names, defaults []string) (Columns[T], error) {
	switch {
	case all:
		return cs, nil
	case len(names) > 0:
		return cs.Select(names)
	default:
		return cs.Select(defaults)
	}
}

// Cell is one table cell.
type Cell struct {
	Text string
	Kind Kind
}

// Table is a projected tabular view.
type Table struct {
	Headers []string
	Aligns  []Align
	Rows    [][]Cell
}

// Project builds a Table from items and selected columns.
func Project[T any](items []T, cols Columns[T]) Table {
	t := Table{
		Headers: make([]string, len(cols)),
		Aligns:  make([]Align, len(cols)),
		Rows:    make([][]Cell, 0, len(items)),
	}
	for i, c := range cols {
		t.Headers[i] = c.Header
		t.Aligns[i] = c.Align
	}
	for _, item := range items {
		row := make([]Cell, len(cols))
		for i, c := range cols {
			row[i] = Cell{Text: c.Value(item)}
			if c.Kind != nil {
				row[i].Kind = c.Kind(item)
			}
		}
		t.Rows = append(t.Rows, row)
	}
	return t
}
