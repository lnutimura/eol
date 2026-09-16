// Package eol is a client for the endoflife.date v1 API.
package eol

import (
	"encoding/json"
	"strings"
	"time"
)

// Date is a calendar date that distinguishes omitted/null from a real value.
type Date struct {
	Time  time.Time
	Valid bool
}

func (d Date) String() string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// MarshalJSON encodes a valid date as YYYY-MM-DD and an invalid date as null.
func (d Date) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// UnmarshalJSON accepts a YYYY-MM-DD string or null.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == `""` || s == "" {
		*d = Date{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		*d = Date{}
		return nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		t, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return err
		}
	}
	d.Time = t
	d.Valid = true
	return nil
}

// Response is the API envelope shared by every endpoint.
type Response[T any] struct {
	SchemaVersion string     `json:"schema_version"`
	GeneratedAt   time.Time  `json:"generated_at"`
	LastModified  *time.Time `json:"last_modified,omitempty"`
	Total         *int       `json:"total,omitempty"`
	Result        T          `json:"result"`
}

// Uri is a named link to an API resource.
type Uri struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// Identifier is a product identifier such as a purl, CPE, or Repology id.
type Identifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// IdentifierRef pairs an identifier value with its product.
type IdentifierRef struct {
	Identifier string `json:"identifier"`
	Product    Uri    `json:"product"`
}

// ProductSummary is the compact product record returned by list endpoints.
type ProductSummary struct {
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases"`
	Label    string   `json:"label"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	URI      string   `json:"uri"`
}

// ProductDetails is the full product record, including releases.
type ProductDetails struct {
	Name           string               `json:"name"`
	Aliases        []string             `json:"aliases"`
	Label          string               `json:"label"`
	Category       string               `json:"category"`
	Tags           []string             `json:"tags"`
	VersionCommand *string              `json:"versionCommand"`
	Identifiers    []Identifier         `json:"identifiers"`
	Labels         ProductSupportLabels `json:"labels"`
	Links          ProductLinks         `json:"links"`
	Releases       []ProductRelease     `json:"releases"`
}

// ProductSupportLabels holds the human labels for each support phase.
type ProductSupportLabels struct {
	Eoas         *string `json:"eoas"`
	Discontinued *string `json:"discontinued"`
	Eol          string  `json:"eol"`
	Eoes         *string `json:"eoes"`
}

// ProductLinks holds product URLs.
type ProductLinks struct {
	Icon          *string `json:"icon"`
	HTML          string  `json:"html"`
	ReleasePolicy *string `json:"releasePolicy"`
}

// ProductVersion is the latest version documented for a release cycle.
type ProductVersion struct {
	Name string  `json:"name"`
	Date Date    `json:"date"`
	Link *string `json:"link"`
}

// ProductRelease is a product release cycle.
type ProductRelease struct {
	Name             string             `json:"name"`
	Codename         *string            `json:"codename"`
	Label            string             `json:"label"`
	ReleaseDate      Date               `json:"releaseDate"`
	IsLts            bool               `json:"isLts"`
	LtsFrom          Date               `json:"ltsFrom"`
	IsEoas           *bool              `json:"isEoas,omitempty"`
	EoasFrom         Date               `json:"eoasFrom"`
	IsEol            bool               `json:"isEol"`
	EolFrom          Date               `json:"eolFrom"`
	IsDiscontinued   *bool              `json:"isDiscontinued,omitempty"`
	DiscontinuedFrom Date               `json:"discontinuedFrom"`
	IsEoes           *bool              `json:"isEoes"`
	EoesFrom         Date               `json:"eoesFrom"`
	IsMaintained     bool               `json:"isMaintained"`
	Latest           *ProductVersion    `json:"latest"`
	Custom           map[string]*string `json:"custom"`
}

// LatestCycle returns the newest release cycle, if any.
func (p ProductDetails) LatestCycle() *ProductRelease {
	if len(p.Releases) == 0 {
		return nil
	}
	return &p.Releases[0]
}
