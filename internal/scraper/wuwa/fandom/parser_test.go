package fandom

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestParseHTMLCodesRewardsAndDuration(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	page := `<table><tr><th>Used?</th><th>Code</th><th>Server</th><th>Rewards</th><th>Duration</th></tr><tr><td></td><td><code>WUTHERINGGIFT</code></td><td>All</td><td><div class="card-container"><img alt="Astrite"><span class="card-text">50</span></div><div class="card-container"><img alt="Shell Credit"><span class="card-text">10,000</span></div></td><td>Valid until: Unknown</td></tr><tr><td></td><td><code>OLDCODE</code></td><td>All</td><td><div class="card-container"><img alt="Astrite"><span class="card-text">100</span></div></td><td>Valid until: July 1, 2026</td></tr><tr><td></td><td><code>CNCODE</code></td><td>China</td><td>Astrite</td><td>Valid until: Unknown</td></tr></table>`
	candidates, err := parseHTML(page, "test", now, 10)
	if err != nil || len(candidates) != 2 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Status != domain.StatusActive || candidates[0].GameSlug != "wuwa" || candidates[0].Rewards[1] != "Shell Credit ×10,000" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
	if candidates[1].Status != domain.StatusExpired || candidates[1].ExpiresAt == nil {
		t.Fatalf("candidate=%+v", candidates[1])
	}
}
