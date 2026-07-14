package mobalytics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectUsesFetchURLAndPageProvenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("user agent=%q", request.Header.Get("User-Agent"))
		}
		_, _ = response.Write([]byte("## Endfield Active Promo Codes\n| Latest Codes | Rewards |\n| --- | --- |\n| ENDFIELDGIFT | * T-Creds x13,000 |\n## How to Redeem"))
	}))
	defer server.Close()

	candidates, err := (&Source{Client: server.Client(), FetchURL: server.URL, PageURL: "https://example.com/codes", UserAgent: "test-agent"}).Collect(t.Context())
	if err != nil || len(candidates) != 1 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].SourceURL != "https://example.com/codes" {
		t.Fatalf("source URL=%q", candidates[0].SourceURL)
	}
}
