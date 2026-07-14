package game8

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestParseHTMLUsesOnlyActiveTable(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	page := `<h3 id="hm_1">All Active Redeem Codes</h3><table class="a-table"><tr><th>Redeem Code</th><th>Rewards</th></tr><tr><td><input class="a-clipboard__textInput" value="NTEGIFT"><span>Expiry Date: 05/13/2026 (Expiry date passed but code is still active)</span></td><td><div class="align">・ Annulith x50</div><div class="align">・ Fons x20000</div></td></tr></table><h2 id="hl_2">Expired Codes</h2><h3 id="hm_2">Expired</h3><table class="a-table"><tr><th>Code</th><th>Rewards</th></tr><tr><td><input class="a-clipboard__textInput" value="OLDCODE"></td><td>Annulith x1</td></tr></table>`
	candidates, err := parseHTML(page, "test", now)
	if err != nil || len(candidates) != 1 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	candidate := candidates[0]
	if candidate.Code != "NTEGIFT" || candidate.GameSlug != "nte" || candidate.Status != domain.StatusActive || candidate.ExpiresAt != nil {
		t.Fatalf("candidate=%+v", candidate)
	}
	if candidate.Rewards[0] != "Annulith ×50" || candidate.Rewards[1] != "Fons ×20,000" {
		t.Fatalf("rewards=%+v", candidate.Rewards)
	}
}
