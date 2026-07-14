package sportskeeda

import (
	"testing"
	"time"
)

func TestParseHTMLActiveCodesAndRewards(t *testing.T) {
	page := `<h2>All active Honkai Star Rail 4.0 redeem codes</h2>
<p>Active codes:</p>
<ul>
<li><strong>4T28RJM4MS3B</strong>: Stellar Jade x50, Credit x10,000</li>
<li><strong>AT45Q:</strong> Stellar Jade x100, Credit 50,000x</li>
<li><strong>STARRAILGIFT</strong>: Stellar Jade x100, Traveler's Guide x4, Bottled Soda x5, 50,000 Credit</li>
</ul>
<h2>How to redeem codes</h2>
<ul><li><strong>DECOY123</strong>: Credit x1</li></ul>`

	observedAt := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	candidates, err := parseHTML(page, "https://example.com/codes", observedAt)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 3 {
		t.Fatalf("candidates=%d", len(candidates))
	}
	if candidates[0].Code != "4T28RJM4MS3B" || candidates[0].GameSlug != "starrail" || candidates[0].SourceID != "starrail-sportskeeda" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
	if got := candidates[0].Rewards; len(got) != 2 || got[0] != "Stellar Jade ×50" || got[1] != "Credit ×10,000" {
		t.Fatalf("rewards=%q", got)
	}
	if got := candidates[1].Rewards; len(got) != 2 || got[1] != "Credit ×50,000" {
		t.Fatalf("rewards=%q", got)
	}
	if got := candidates[2].Rewards; len(got) != 4 || got[3] != "Credit ×50,000" {
		t.Fatalf("rewards=%q", got)
	}
}

func TestNormalizeRewardFormats(t *testing.T) {
	for input, want := range map[string]string{
		"Stellar Jade x50": "Stellar Jade ×50",
		"Credit 50,000x":   "Credit ×50,000",
		"50,000 Credit":    "Credit ×50,000",
	} {
		if got := normalizeReward(input); got != want {
			t.Errorf("normalizeReward(%q)=%q want %q", input, got, want)
		}
	}
}

func TestParseHTMLRequiresExactActiveSection(t *testing.T) {
	if _, err := parseHTML(`<h2>Expired Honkai Star Rail redeem codes</h2><ul><li><strong>OLD123</strong>: Credit x1</li></ul>`, "test", time.Now()); err == nil {
		t.Fatal("expected missing active section error")
	}
}
