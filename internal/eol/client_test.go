package eol

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(Options{
		BaseURL:   srv.URL,
		CacheDir:  t.TempDir(),
		CacheTTL:  time.Hour,
		UserAgent: "eol/test",
		Timeout:   5 * time.Second,
	})
}

func envelope(result any) []byte {
	b, err := json.Marshal(map[string]any{
		"schema_version": "1.2.1",
		"generated_at":   "2026-09-16T00:00:00Z",
		"total":          1,
		"result":         result,
	})
	if err != nil {
		panic(err)
	}
	return b
}

func TestGetOK(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "eol/test" {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(envelope([]map[string]string{{"name": "products", "uri": "https://example.com/products"}}))
	}))
	resp, err := c.Index(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Result) != 1 || resp.Result[0].Name != "products" {
		t.Fatalf("result = %+v", resp.Result)
	}
}

func TestGetFollows301(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/products/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/products/new", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/products/new", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(envelope(map[string]any{
			"name": "new", "label": "New", "aliases": []string{}, "category": "os",
			"tags": []string{"os"}, "identifiers": []any{}, "labels": map[string]any{"eol": "EOL"},
			"links": map[string]any{"html": "https://example.com"}, "releases": []any{},
		}))
	})
	c := testClient(t, mux)
	resp, err := c.GetProduct(context.Background(), "old")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result.Name != "new" {
		t.Fatalf("got %+v", resp.Result)
	}
}

func TestGetNotFoundHTML(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "<html>not found</html>")
	}))
	_, err := c.GetProduct(context.Background(), "typo")
	nf, ok := err.(*NotFoundError)
	if !ok {
		t.Fatalf("got %T %v", err, err)
	}
	if nf.Kind != "product" || nf.Name != "typo" {
		t.Fatalf("got %+v", nf)
	}
}

func TestGetNonJSON(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "hello")
	}))
	_, err := c.Index(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGet304FromCache(t *testing.T) {
	var hits atomic.Int32
	body := envelope([]map[string]string{{"name": "os", "uri": "https://example.com/os"}})
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if inm := r.Header.Get("If-None-Match"); inm == `"abc"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"abc"`)
		_, _ = w.Write(body)
	}))
	c.CacheTTL = 0

	first, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
	if len(first.Result) != 1 || first.Result[0].Name != second.Result[0].Name {
		t.Fatalf("first=%+v second=%+v", first.Result, second.Result)
	}
}

func TestGetFreshCacheSkipsNetwork(t *testing.T) {
	var hits atomic.Int32
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"x"`)
		_, _ = w.Write(envelope([]map[string]string{{"name": "os", "uri": "https://example.com/os"}}))
	}))
	if _, err := c.ListCategories(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListCategories(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}
}

func TestGetDisableCache(t *testing.T) {
	var hits atomic.Int32
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("If-None-Match") != "" {
			t.Error("If-None-Match should not be sent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(envelope([]map[string]string{{"name": "os", "uri": "https://example.com/os"}}))
	}))
	c.DisableCache = true
	if _, err := c.ListCategories(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListCategories(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestGet429RetryAfter(t *testing.T) {
	var hits atomic.Int32
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(envelope([]map[string]string{{"name": "os", "uri": "https://example.com/os"}}))
	}))
	c.DisableCache = true
	resp, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
	if len(resp.Result) != 1 {
		t.Fatalf("result = %+v", resp.Result)
	}
}

func TestGet429Exhausted(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	c.DisableCache = true
	_, err := c.ListCategories(context.Background())
	if _, ok := err.(*RateLimitError); !ok {
		t.Fatalf("got %T %v", err, err)
	}
}
