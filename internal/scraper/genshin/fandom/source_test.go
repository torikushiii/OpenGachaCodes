package fandom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSourceCollect(t *testing.T) {
	fixture, err := os.ReadFile("testdata/promotional_code.json")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("action"); got != "parse" {
			t.Errorf("action = %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "Promotional_Code" {
			t.Errorf("page = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Errorf("User-Agent = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	source := Source{Client: server.Client(), APIURL: server.URL, UserAgent: "test-agent", Now: func() time.Time { return now }}
	codes, err := source.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 2 {
		t.Fatalf("got %d codes, want 2", len(codes))
	}
	if codes[0].Code != "GENSHINGIFT" || len(codes[0].Rewards) != 2 {
		t.Fatalf("unexpected first code: %#v", codes[0])
	}
	if codes[0].RevisionID != 12345 || !codes[0].ObservedAt.Equal(now) {
		t.Fatalf("missing observation metadata: %#v", codes[0])
	}
	if codes[1].Status != "expired" {
		t.Fatalf("second status = %q", codes[1].Status)
	}
	for _, code := range codes {
		if code.Code == "YuanShen" {
			t.Fatal("China-only code was collected")
		}
	}
}

func TestSourceCollectErrors(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
	}{
		{name: "http status", code: http.StatusTooManyRequests, body: `{}`},
		{name: "malformed JSON", code: http.StatusOK, body: `{`},
		{name: "missing HTML", code: http.StatusOK, body: `{"parse":{"text":""}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			source := Source{Client: server.Client(), APIURL: server.URL, UserAgent: "test-agent"}
			if _, err := source.Collect(context.Background()); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSourceRequiresUserAgent(t *testing.T) {
	source := Source{APIURL: "http://example.invalid"}
	if _, err := source.Collect(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
