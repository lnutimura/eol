package eol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateUnmarshalNull(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte("null"), &d); err != nil {
		t.Fatal(err)
	}
	if d.Valid {
		t.Fatalf("null date should be invalid, got %v", d.Time)
	}
}

func TestDateUnmarshalValue(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2031-05-29"`), &d); err != nil {
		t.Fatal(err)
	}
	if !d.Valid {
		t.Fatal("expected valid date")
	}
	want := time.Date(2031, 5, 29, 0, 0, 0, 0, time.UTC)
	if !d.Time.Equal(want) {
		t.Fatalf("got %v want %v", d.Time, want)
	}
}

func TestDateMarshalNull(t *testing.T) {
	b, err := json.Marshal(Date{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Fatalf("got %s", b)
	}
}

func TestProductReleaseOptionalFields(t *testing.T) {
	const raw = `{
		"name": "26.04",
		"codename": "Resolute Raccoon",
		"label": "26.04 LTS",
		"releaseDate": "2026-04-23",
		"isLts": true,
		"ltsFrom": null,
		"isEoas": false,
		"eoasFrom": "2031-05-29",
		"isEol": false,
		"eolFrom": "2031-05-29",
		"isEoes": false,
		"eoesFrom": "2036-04-23",
		"isMaintained": true,
		"latest": {"name": "26.04.1", "date": "2026-08-31", "link": "https://example.com"},
		"custom": null
	}`
	var r ProductRelease
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.IsEoas == nil || *r.IsEoas {
		t.Fatalf("isEoas: %+v", r.IsEoas)
	}
	if !r.EoasFrom.Valid || r.EoasFrom.String() != "2031-05-29" {
		t.Fatalf("eoasFrom: %+v", r.EoasFrom)
	}
	if r.LtsFrom.Valid {
		t.Fatal("ltsFrom should be null")
	}
	if r.Custom != nil {
		t.Fatalf("custom: %+v", r.Custom)
	}
	if r.Latest == nil || r.Latest.Name != "26.04.1" {
		t.Fatalf("latest: %+v", r.Latest)
	}
}
