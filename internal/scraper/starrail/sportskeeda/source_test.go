package sportskeeda

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCollectUsesBrowserAndProjectUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		userAgent := request.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "Mozilla/5.0") || !strings.Contains(userAgent, "OpenGachaCodes") {
			t.Errorf("user agent=%q", userAgent)
		}
		if request.Header.Get("Accept-Language") == "" {
			t.Error("missing Accept-Language")
		}
		_, _ = response.Write([]byte(`<h2>All active Honkai Star Rail 4.0 redeem codes</h2><ul><li><strong>CODE123</strong>: Stellar Jade x50</li></ul>`))
	}))
	defer server.Close()

	observedAt := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	candidates, err := (&Source{Client: server.Client(), PageURL: server.URL, UserAgent: "OpenGachaCodes (+https://github.com/torikushiii/OpenGachaCodes)", Now: func() time.Time { return observedAt }}).Collect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].ObservedAt != observedAt || candidates[0].SourceURL != server.URL {
		t.Fatalf("candidates=%+v", candidates)
	}
}

func TestCollectRejectsChallengeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = response.Write([]byte("Human Verification"))
	}))
	defer server.Close()

	_, err := (&Source{Client: server.Client(), PageURL: server.URL, UserAgent: "test-agent"}).Collect(t.Context())
	if err == nil || !strings.Contains(err.Error(), "405") {
		t.Fatalf("error=%v", err)
	}
}
