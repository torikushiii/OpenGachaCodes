package fandom

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestParseHTMLCodesRegionsAndDuration(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	page := `<table><tr><th>Code</th><th>Server</th><th>Rewards</th><th>Duration</th></tr><tr><td><code>ZZZSTEAM</code><a>Quick Redeem</a></td><td>All</td><td><span class="item">Polychrome ×60</span><span class="item">Denny ×6,666</span></td><td>Valid until: July 29, 2026</td></tr><tr><td><code>ASIAONLY</code></td><td>Asia</td><td><span class="item">Denny ×20,000</span></td><td>Valid until: Unknown</td></tr><tr><td><code>OLDCODE</code></td><td>All</td><td>Denny</td><td>Expired: July 1, 2026</td></tr><tr><td><code>CNCODE</code></td><td>China</td><td>Denny</td><td>Valid until: Unknown</td></tr></table>`
	candidates, err := parseHTML(page, "test", now, 10)
	if err != nil || len(candidates) != 3 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Status != domain.StatusActive || candidates[0].ExpiresAt == nil || candidates[0].Region != "global" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
	if candidates[1].Status != domain.StatusActive || candidates[1].Region != "global" {
		t.Fatalf("candidate=%+v", candidates[1])
	}
	if candidates[2].Status != domain.StatusExpired {
		t.Fatalf("candidate=%+v", candidates[2])
	}
}
