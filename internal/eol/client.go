package eol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultBaseURL is the production endoflife.date v1 API.
	DefaultBaseURL = "https://endoflife.date/api/v1"
	// DefaultTimeout is the default HTTP and context timeout.
	DefaultTimeout = 10 * time.Second
	// DefaultTTL is how long a cached response is served without revalidation.
	DefaultTTL = time.Hour

	maxAttempts   = 3
	maxRetryWait  = 30 * time.Second
	defaultUAName = "eol"
)

// Options configures a Client.
type Options struct {
	BaseURL      string
	HTTPClient   *http.Client
	UserAgent    string
	CacheDir     string
	CacheTTL     time.Duration
	DisableCache bool
	Timeout      time.Duration
}

// Client talks to the endoflife.date v1 API.
type Client struct {
	BaseURL      string
	HTTPClient   *http.Client
	UserAgent    string
	CacheDir     string
	CacheTTL     time.Duration
	DisableCache bool

	mu sync.Mutex
}

// New returns a Client with defaults applied.
func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	base := opts.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	ttl := opts.CacheTTL
	if ttl < 0 {
		ttl = DefaultTTL
	}
	ua := opts.UserAgent
	if ua == "" {
		ua = defaultUAName + "/dev"
	}
	return &Client{
		BaseURL:      strings.TrimRight(base, "/"),
		HTTPClient:   httpClient,
		UserAgent:    ua,
		CacheDir:     opts.CacheDir,
		CacheTTL:     ttl,
		DisableCache: opts.DisableCache,
	}
}

func (c *Client) ttl() time.Duration {
	if c.CacheTTL < 0 {
		return DefaultTTL
	}
	return c.CacheTTL
}

// Index lists the main API endpoints.
func (c *Client) Index(ctx context.Context) (Response[[]Uri], error) {
	return get[[]Uri](ctx, c, "/", "index", "")
}

// ListProducts returns a summary of every product.
func (c *Client) ListProducts(ctx context.Context) (Response[[]ProductSummary], error) {
	return get[[]ProductSummary](ctx, c, "/products", "products", "")
}

// ListProductsFull returns every product with full details.
func (c *Client) ListProductsFull(ctx context.Context) (Response[[]ProductDetails], error) {
	return get[[]ProductDetails](ctx, c, "/products/full", "products", "")
}

// GetProduct returns one product, including its release cycles.
func (c *Client) GetProduct(ctx context.Context, name string) (Response[ProductDetails], error) {
	return get[ProductDetails](ctx, c, "/products/"+url.PathEscape(name), "product", name)
}

// GetRelease returns one release cycle for a product.
func (c *Client) GetRelease(ctx context.Context, product, release string) (Response[ProductRelease], error) {
	path := "/products/" + url.PathEscape(product) + "/releases/" + url.PathEscape(release)
	return get[ProductRelease](ctx, c, path, "release", product+"@"+release)
}

// GetLatestRelease returns the latest release cycle for a product.
func (c *Client) GetLatestRelease(ctx context.Context, product string) (Response[ProductRelease], error) {
	path := "/products/" + url.PathEscape(product) + "/releases/latest"
	return get[ProductRelease](ctx, c, path, "product", product)
}

// ListCategories lists all product categories.
func (c *Client) ListCategories(ctx context.Context) (Response[[]Uri], error) {
	return get[[]Uri](ctx, c, "/categories", "categories", "")
}

// ListProductsByCategory returns product summaries in a category.
func (c *Client) ListProductsByCategory(ctx context.Context, category string) (Response[[]ProductSummary], error) {
	return get[[]ProductSummary](ctx, c, "/categories/"+url.PathEscape(category), "category", category)
}

// ListTags lists all product tags.
func (c *Client) ListTags(ctx context.Context) (Response[[]Uri], error) {
	return get[[]Uri](ctx, c, "/tags", "tags", "")
}

// ListProductsByTag returns product summaries with the given tag.
func (c *Client) ListProductsByTag(ctx context.Context, tag string) (Response[[]ProductSummary], error) {
	return get[[]ProductSummary](ctx, c, "/tags/"+url.PathEscape(tag), "tag", tag)
}

// ListIdentifierTypes lists known identifier types such as purl and cpe.
func (c *Client) ListIdentifierTypes(ctx context.Context) (Response[[]Uri], error) {
	return get[[]Uri](ctx, c, "/identifiers", "identifiers", "")
}

// ListIdentifiers returns identifiers of a given type and their products.
func (c *Client) ListIdentifiers(ctx context.Context, identType string) (Response[[]IdentifierRef], error) {
	return get[[]IdentifierRef](ctx, c, "/identifiers/"+url.PathEscape(identType), "identifier type", identType)
}

func get[T any](ctx context.Context, c *Client, path, kind, name string) (Response[T], error) {
	var zero Response[T]
	rawURL := c.BaseURL + "/" + strings.TrimLeft(path, "/")
	key := cacheKey(rawURL)

	var cached *cacheEntry
	if !c.DisableCache {
		c.mu.Lock()
		cached = c.readCache(key)
		c.mu.Unlock()
		if cached != nil && c.ttl() > 0 && time.Since(cached.CachedAt) < c.ttl() {
			return decode[T](cached.Body)
		}
	}

	var retryAfter string
	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return zero, err
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json")
		if !c.DisableCache && cached != nil && cached.ETag != "" {
			req.Header.Set("If-None-Match", cached.ETag)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return zero, fmt.Errorf("request %s: %w", rawURL, err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return zero, fmt.Errorf("read response: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			if err := ensureJSON(resp.Header.Get("Content-Type")); err != nil {
				return zero, err
			}
			if !c.DisableCache {
				c.mu.Lock()
				c.writeCache(key, cacheEntry{
					ETag:     resp.Header.Get("ETag"),
					Body:     json.RawMessage(body),
					CachedAt: time.Now(),
				})
				c.mu.Unlock()
			}
			return decode[T](body)
		case http.StatusNotModified:
			if cached == nil {
				return zero, fmt.Errorf("received 304 without a cached body")
			}
			cached.CachedAt = time.Now()
			if etag := resp.Header.Get("ETag"); etag != "" {
				cached.ETag = etag
			}
			if !c.DisableCache {
				c.mu.Lock()
				c.writeCache(key, *cached)
				c.mu.Unlock()
			}
			return decode[T](cached.Body)
		case http.StatusNotFound:
			return zero, &NotFoundError{Kind: kind, Name: name}
		case http.StatusTooManyRequests:
			retryAfter = resp.Header.Get("Retry-After")
			if attempt == maxAttempts-1 {
				return zero, &RateLimitError{RetryAfter: retryAfter}
			}
			wait := parseRetryAfter(retryAfter)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(wait):
			}
		default:
			return zero, fmt.Errorf("unexpected status %s from %s", resp.Status, rawURL)
		}
	}
	return zero, &RateLimitError{RetryAfter: retryAfter}
}

func decode[T any](body []byte) (Response[T], error) {
	var resp Response[T]
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, fmt.Errorf("decode response: %w", err)
	}
	return resp, nil
}

func ensureJSON(contentType string) error {
	if contentType == "" {
		return nil
	}
	media, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("non-JSON response (content-type %q)", contentType)
	}
	if media != "application/json" {
		return fmt.Errorf("non-JSON response (content-type %q)", contentType)
	}
	return nil
}

func parseRetryAfter(h string) time.Duration {
	h = strings.TrimSpace(h)
	if h == "" {
		return time.Second
	}
	if n, err := strconv.Atoi(h); err == nil {
		if n < 0 {
			n = 0
		}
		d := time.Duration(n) * time.Second
		if d > maxRetryWait {
			return maxRetryWait
		}
		return d
	}
	t, err := http.ParseTime(h)
	if err != nil {
		return time.Second
	}
	d := time.Until(t)
	if d < 0 {
		return 0
	}
	if d > maxRetryWait {
		return maxRetryWait
	}
	return d
}
