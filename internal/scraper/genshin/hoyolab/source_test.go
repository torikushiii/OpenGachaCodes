package hoyolab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSourceCollect(t *testing.T) {
	fixture, err := os.ReadFile("testdata/material_guide.json")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("game_id") != "2" {
			t.Errorf("game_id=%q", r.URL.Query().Get("game_id"))
		}
		headers := map[string]string{
			"User-Agent":        "test-agent",
			"x-rpc-app_version": "4.8.0",
			"x-rpc-client_type": "4",
			"x-rpc-language":    "en-us",
			"Referer":           DefaultReferer,
		}
		for name, want := range headers {
			if got := r.Header.Get(name); got != want {
				t.Errorf("%s=%q, want %q", name, got, want)
			}
		}
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	source := Source{
		Client: server.Client(), APIURL: server.URL + "?game_id=2", UserAgent: "test-agent",
		Now: func() time.Time { return now },
	}
	codes, err := source.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 1 {
		t.Fatalf("got %d codes, want 1", len(codes))
	}
	code := codes[0]
	if code.Code != "OFFICIAL123" || code.Region != "global" || code.Authority != "official" || !code.ObservedAt.Equal(now) {
		t.Fatalf("unexpected candidate: %#v", code)
	}
	if len(code.Rewards) != 5 {
		t.Fatalf("rewards=%#v", code.Rewards)
	}
	if !strings.Contains(strings.Join(code.Rewards, "|"), "Unknown reward (new_reward_hash) ×1") {
		t.Fatalf("unknown reward was not preserved: %#v", code.Rewards)
	}
}

func TestSourceErrors(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
	}{
		{name: "http", code: http.StatusTooManyRequests, body: `{}`},
		{name: "json", code: http.StatusOK, body: `{`},
		{name: "api", code: http.StatusOK, body: `{"retcode":-1,"message":"failed"}`},
		{name: "no codes", code: http.StatusOK, body: `{"retcode":0,"data":{"modules":[{}]}}`},
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
	if _, err := (&Source{APIURL: "http://example.invalid"}).Collect(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
