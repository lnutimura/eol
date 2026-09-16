package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type row struct {
	Name string
	Age  int
}

func testCols() Columns[row] {
	return Columns[row]{
		{Name: "name", Header: "Name", Value: func(r row) string { return r.Name }},
		{Name: "age", Header: "Age", Value: func(r row) string { return "21" }, Align: AlignRight},
	}
}

func TestSelectUnknown(t *testing.T) {
	_, err := testCols().Select([]string{"Nmae"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `"Nmae"`) {
		t.Fatalf("error should mention the bad name: %v", err)
	}
	if !strings.Contains(err.Error(), "name") || !strings.Contains(err.Error(), "age") {
		t.Fatalf("error should list valid columns: %v", err)
	}
}

func TestSelectCaseInsensitive(t *testing.T) {
	cols, err := testCols().Select([]string{"NAME", "Age"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 || cols[0].Name != "name" || cols[1].Name != "age" {
		t.Fatalf("got %+v", cols)
	}
}

func TestPickAll(t *testing.T) {
	cols, err := testCols().Pick(true, []string{"name"}, []string{"name"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 {
		t.Fatalf("len = %d", len(cols))
	}
}

func TestRenderFormats(t *testing.T) {
	items := []row{{Name: "ubuntu", Age: 21}}
	cols, err := testCols().Select([]string{"name", "age"})
	if err != nil {
		t.Fatal(err)
	}
	view := View{
		Data: items,
		Meta: Meta{SchemaVersion: "1.2.1", GeneratedAt: time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)},
		Table: func() Table {
			return Project(items, cols)
		},
	}

	t.Run("table", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatTable, Color: ColorNever}, view); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if !strings.Contains(out, "NAME") || !strings.Contains(out, "ubuntu") {
			t.Fatalf("table = %q", out)
		}
		if strings.Contains(out, "╭") || strings.Contains(out, "┌") {
			t.Fatalf("plain table should be borderless: %q", out)
		}
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatCSV}, view); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "Name,Age\nubuntu,21\n" {
			t.Fatalf("csv = %q", got)
		}
	})

	t.Run("tsv", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatTSV}, view); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "Name\tAge\nubuntu\t21\n" {
			t.Fatalf("tsv = %q", got)
		}
	})

	t.Run("markdown", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatMarkdown}, view); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, "| Name | Age |") || !strings.Contains(got, "| ubuntu | 21 |") {
			t.Fatalf("markdown = %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatJSON}, view); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, `"Name": "ubuntu"`) {
			t.Fatalf("json = %q", got)
		}
		if strings.Contains(got, "schema_version") {
			t.Fatalf("json without --meta should be the payload: %q", got)
		}
	})

	t.Run("json-meta", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatJSON, Meta: true}, view); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, `"schema_version": "1.2.1"`) {
			t.Fatalf("json meta = %q", got)
		}
	})

	t.Run("yaml", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Print(&buf, Options{Format: FormatYAML}, view); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, "Name: ubuntu") {
			t.Fatalf("yaml = %q", got)
		}
	})
}

func TestDetailThenTable(t *testing.T) {
	var buf bytes.Buffer
	err := Print(&buf, Options{Format: FormatTable, Color: ColorNever}, View{
		Detail: []Field{{Label: "Name", Value: "ubuntu"}},
		Table: func() Table {
			return Project([]row{{Name: "22.04"}}, Columns[row]{
				{Name: "name", Header: "Name", Value: func(r row) string { return r.Name }},
			})
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Name:") || !strings.Contains(out, "ubuntu") || !strings.Contains(out, "22.04") {
		t.Fatalf("got %q", out)
	}
}

func TestParseFormat(t *testing.T) {
	f, err := ParseFormat("MD")
	if err != nil || f != FormatMarkdown {
		t.Fatalf("got %q %v", f, err)
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Fatal("expected error")
	}
}
