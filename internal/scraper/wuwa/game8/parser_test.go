package game8

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestParseHTMLUsesAllTablesInActiveSection(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	page := `<h2 id="hl_1">Wuthering Waves Codes</h2><h3 id="hm_1">Collaboration</h3><table class="a-table"><tr><th>Code</th></tr><tr><td><input class="a-clipboard__textInput" value="COLLAB123"><b>Expiry:</b> August 19, 2026<div class="align">Livery x1</div></td></tr></table><h3 id="hm_2">All Active Codes</h3><table class="a-table"><tr><th>Codes</th></tr><tr><td><input class="a-clipboard__textInput" value="WUTHERINGGIFT"><div class="align">Astrite x50</div><div class="align">Shell Credit x15000</div></td></tr></table><h2 id="hl_4">Expired</h2><table class="a-table"><tr><td><input class="a-clipboard__textInput" value="OLDCODE"><div class="align">Astrite x1</div></td></tr></table>`
	candidates, err := parseHTML(page, "test", now)
	if err != nil || len(candidates) != 2 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Code != "COLLAB123" || candidates[0].ExpiresAt == nil || candidates[0].Status != domain.StatusActive {
		t.Fatalf("candidate=%+v", candidates[0])
	}
	if candidates[1].Code != "WUTHERINGGIFT" || candidates[1].GameSlug != "wuwa" || candidates[1].Rewards[1] != "Shell Credit ×15,000" {
		t.Fatalf("candidate=%+v", candidates[1])
	}
}
