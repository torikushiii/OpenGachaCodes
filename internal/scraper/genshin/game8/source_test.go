package game8

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSourceCollectsOnlyActiveSection(t *testing.T) {
	fixture, err := os.ReadFile("testdata/active_codes.html")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test-agent" || r.Header.Get("Accept") != "text/html" {
			t.Errorf("unexpected headers: %#v", r.Header)
		}
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	source := Source{Client: server.Client(), PageURL: server.URL, UserAgent: "test-agent", Now: func() time.Time { return now }}
	codes, err := source.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 1 {
		t.Fatalf("got %d codes: %#v", len(codes), codes)
	}
	code := codes[0]
	if code.Code != "ACTIVE123" || code.SourceID != "genshin-game8" || code.Authority != "community" || !code.ObservedAt.Equal(now) {
		t.Fatalf("unexpected candidate: %#v", code)
	}
	if len(code.Rewards) != 2 || code.Rewards[0] != "Primogem ×60" || code.Rewards[1] != "Adventurer's Experience ×5" {
		t.Fatalf("rewards=%#v", code.Rewards)
	}
}

func TestSourceErrors(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
	}{
		{name: "http", code: http.StatusForbidden, body: "blocked"},
		{name: "missing section", code: http.StatusOK, body: "<html></html>"},
		{name: "missing table", code: http.StatusOK, body: `<h3 id="hm_1">Active</h3><h3>Next</h3>`},
		{name: "missing headers", code: http.StatusOK, body: `<h3 id="hm_1">Active</h3><table class="a-table"><tr><th>Other</th></tr></table>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			source := Source{Client: server.Client(), PageURL: server.URL, UserAgent: "test-agent"}
			if _, err := source.Collect(context.Background()); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSourceRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxResponseSize+1)))
	}))
	defer server.Close()
	source := Source{Client: server.Client(), PageURL: server.URL, UserAgent: "test-agent"}
	if _, err := source.Collect(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestSourceRequiresUserAgent(t *testing.T) {
	if _, err := (&Source{PageURL: "http://example.invalid"}).Collect(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
