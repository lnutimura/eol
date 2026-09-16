package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lnutimura/eol/internal/eol"
)

func runCLI(t *testing.T, srv *httptest.Server, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewRootCmd("test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{
		"--api-url", srv.URL,
		"--no-cache",
		"--color", "never",
	}, args...))
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func apiServer(t *testing.T, routes map[string]any) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, payload := range routes {
		p, body := path, payload
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			switch b := body.(type) {
			case int:
				if b == http.StatusNotFound {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusNotFound)
					_, _ = io.WriteString(w, "<html>not found</html>")
					return
				}
				w.WriteHeader(b)
				return
			default:
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(body)
			}
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func wrap(result any) map[string]any {
	return map[string]any{
		"schema_version": "1.2.1",
		"generated_at":   "2026-09-16T00:00:00Z",
		"total":          1,
		"result":         result,
	}
}

func wrapProduct(result any) map[string]any {
	return map[string]any{
		"schema_version": "1.2.1",
		"generated_at":   "2026-09-16T00:00:00Z",
		"last_modified":  "2026-09-01T00:00:00Z",
		"result":         result,
	}
}

func TestProductListJSON(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products": wrap([]map[string]any{{
			"name": "ubuntu", "aliases": []string{}, "label": "Ubuntu",
			"category": "os", "tags": []string{"os"}, "uri": "https://example.com/ubuntu",
		}}),
	})
	out, _, err := runCLI(t, srv, "-o", "json", "product", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"name": "ubuntu"`) {
		t.Fatalf("out = %s", out)
	}
}

func TestProductListCSVCategory(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/categories/os": wrap([]map[string]any{{
			"name": "ubuntu", "aliases": []string{}, "label": "Ubuntu",
			"category": "os", "tags": []string{"os"}, "uri": "https://example.com/ubuntu",
		}}),
	})
	out, _, err := runCLI(t, srv, "-o", "csv", "product", "list", "--category", "os")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "ubuntu") {
		t.Fatalf("out = %s", out)
	}
}

func TestProductGetDetailAndReleases(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products/ubuntu": wrapProduct(map[string]any{
			"name": "ubuntu", "label": "Ubuntu", "aliases": []string{"ubuntu-linux"},
			"category": "os", "tags": []string{"canonical", "os"},
			"versionCommand": "lsb_release --release",
			"identifiers":    []map[string]string{{"id": "cpe:/o:canonical:ubuntu_linux", "type": "cpe"}},
			"labels":         map[string]any{"eoas": "Hardware & Maintenance", "discontinued": nil, "eol": "Maintenance", "eoes": "ESM"},
			"links":          map[string]any{"icon": nil, "html": "https://endoflife.date/ubuntu", "releasePolicy": "https://ubuntu.com"},
			"releases": []map[string]any{{
				"name": "26.04", "codename": "Resolute Raccoon", "label": "26.04 LTS",
				"releaseDate": "2026-04-23", "isLts": true, "ltsFrom": nil,
				"isEoas": false, "eoasFrom": "2031-05-29",
				"isEol": false, "eolFrom": "2031-05-29",
				"isEoes": false, "eoesFrom": "2036-04-23",
				"isMaintained": true,
				"latest":       map[string]any{"name": "26.04.1", "date": "2026-08-31", "link": "https://example.com"},
				"custom":       nil,
			}},
		}),
	})
	out, _, err := runCLI(t, srv, "product", "get", "ubuntu")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Ubuntu", "lsb_release --release", "cpe:/o:canonical:ubuntu_linux", "Hardware & Maintenance", "26.04", "2031-05-29"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestProductGetNotFound(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products/typo": http.StatusNotFound,
	})
	_, _, err := runCLI(t, srv, "product", "get", "typo")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `product "typo" not found`) {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "decode") || strings.Contains(err.Error(), "html") {
		t.Fatalf("404 should not be a decode error: %v", err)
	}
}

func TestUnknownColumn(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products": wrap([]map[string]any{{
			"name": "ubuntu", "aliases": []string{}, "label": "Ubuntu",
			"category": "os", "tags": []string{"os"}, "uri": "https://example.com/ubuntu",
		}}),
	})
	_, _, err := runCLI(t, srv, "-c", "Nmae", "product", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `"Nmae"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestReleaseLatestYAML(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products/nodejs/releases/latest": wrap(map[string]any{
			"name": "22", "codename": nil, "label": "22 (LTS)",
			"releaseDate": "2024-04-24", "isLts": true, "ltsFrom": "2024-10-29",
			"isEoas": false, "eoasFrom": "2025-10-21",
			"isEol": false, "eolFrom": "2027-04-30",
			"isEoes": nil, "eoesFrom": nil,
			"isMaintained": true,
			"latest":       map[string]any{"name": "22.19.0", "date": "2026-01-01", "link": nil},
			"custom":       nil,
		}),
	})
	out, _, err := runCLI(t, srv, "-o", "yaml", "release", "latest", "nodejs")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "name: \"22\"") && !strings.Contains(out, "name: 22") {
		t.Fatalf("yaml = %s", out)
	}
}

func TestIndexAndCategoryList(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/":           wrap([]map[string]string{{"name": "products", "uri": "https://example.com/products"}}),
		"/categories": wrap([]map[string]string{{"name": "os", "uri": "https://example.com/os"}}),
	})
	out, _, err := runCLI(t, srv, "-o", "json", "index")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "products") {
		t.Fatalf("index = %s", out)
	}
	out, _, err = runCLI(t, srv, "-o", "csv", "category", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "os") {
		t.Fatalf("categories = %s", out)
	}
}

func TestCheckMaintainedAndEOL(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/products/ubuntu/releases/latest": wrap(map[string]any{
			"name": "26.04", "codename": nil, "label": "26.04",
			"releaseDate": "2026-04-23", "isLts": true, "ltsFrom": nil,
			"isEol": false, "eolFrom": "2031-05-29", "isMaintained": true,
			"latest": nil, "custom": nil,
		}),
		"/products/ubuntu": wrapProduct(map[string]any{
			"name": "ubuntu", "label": "Ubuntu", "aliases": []string{}, "category": "os",
			"tags": []string{"os"}, "identifiers": []any{},
			"labels": map[string]any{"eol": "EOL"}, "links": map[string]any{"html": "https://example.com"},
			"releases": []map[string]any{
				{"name": "22.04", "codename": nil, "label": "22.04", "releaseDate": "2022-04-21",
					"isLts": true, "ltsFrom": nil, "isEol": false, "eolFrom": "2027-04-01", "isMaintained": true,
					"latest": nil, "custom": nil},
				{"name": "18.04", "codename": nil, "label": "18.04", "releaseDate": "2018-04-26",
					"isLts": true, "ltsFrom": nil, "isEol": true, "eolFrom": "2023-05-31", "isMaintained": true,
					"latest": nil, "custom": nil},
			},
		}),
	})

	out, _, err := runCLI(t, srv, "check", "ubuntu")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "maintained") {
		t.Fatalf("out = %s", out)
	}

	out, _, err = runCLI(t, srv, "check", "ubuntu@18.04")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "end-of-life") {
		t.Fatalf("out = %s", out)
	}

	out, _, err = runCLI(t, srv, "check", "ubuntu@22.04.1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "22.04") || !strings.Contains(out, "maintained") {
		t.Fatalf("prefix resolve out = %s", out)
	}
}

func TestCheckStatus(t *testing.T) {
	status, code := checkStatus(eol.ProductRelease{IsEol: true, IsMaintained: true})
	if status != "end-of-life" || code != 2 {
		t.Fatalf("eol+esm: %s %d", status, code)
	}
	status, code = checkStatus(eol.ProductRelease{IsEol: false, IsMaintained: true})
	if status != "maintained" || code != 0 {
		t.Fatalf("supported: %s %d", status, code)
	}
	status, code = checkStatus(eol.ProductRelease{IsEol: false, IsMaintained: false})
	if status != "end-of-life" || code != 2 {
		t.Fatalf("unmaintained: %s %d", status, code)
	}
}

func TestResolveRelease(t *testing.T) {
	releases := []eol.ProductRelease{{Name: "22.04"}, {Name: "22"}, {Name: "18.04"}}
	got, ok := resolveRelease(releases, "22.04.1")
	if !ok || got.Name != "22.04" {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
	if _, ok := resolveRelease(releases, "99"); ok {
		t.Fatal("expected miss")
	}
}

func TestIdentifierGet(t *testing.T) {
	srv := apiServer(t, map[string]any{
		"/identifiers/purl": wrap([]map[string]any{{
			"identifier": "pkg:deb/ubuntu",
			"product":    map[string]string{"name": "ubuntu", "uri": "https://example.com/ubuntu"},
		}}),
	})
	out, _, err := runCLI(t, srv, "-o", "csv", "identifier", "get", "purl")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "pkg:deb/ubuntu") || !strings.Contains(out, "ubuntu") {
		t.Fatalf("out = %s", out)
	}
}

func TestTimeoutFlag(t *testing.T) {
	// Ensure the command still constructs with a tiny timeout; the server is fast.
	srv := apiServer(t, map[string]any{
		"/": wrap([]map[string]string{{"name": "products", "uri": "https://example.com/products"}}),
	})
	_, _, err := runCLI(t, srv, "--timeout", "5s", "-o", "json", "index")
	if err != nil {
		t.Fatal(err)
	}
}
