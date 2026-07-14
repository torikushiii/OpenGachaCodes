package fandom

import (
	"opengachacodes/internal/domain"
	"testing"
	"time"
)

func TestParseHTMLCombinedTableAndExpiry(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	html := `<table><tr><th>Code</th><th>Server</th><th>Rewards</th><th>Duration</th></tr><tr><td>RAIL1234</td><td>Global</td><td><ul><li>Stellar Jade x 50</li><li>Credit x 1000</li></ul></td><td>July 20, 2026</td></tr><tr><td>CN1234</td><td>China</td><td>Credit</td><td>Permanent</td></tr><tr><td>OLD1234</td><td>Global</td><td>Credit</td><td>July 1, 2026</td></tr></table>`
	got, e := parseHTML(html, "test", now, 1)
	if e != nil || len(got) != 2 || got[0].Status != domain.StatusActive || got[0].ExpiresAt == nil || len(got[0].Rewards) != 2 || got[1].Status != domain.StatusExpired {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
func TestDurationAmbiguousIsUnknown(t *testing.T) {
	s, e := duration("Check in game", time.Now())
	if s != domain.StatusUnknown || e != nil {
		t.Fatalf("status=%v expiry=%v", s, e)
	}
}

func TestParseHTMLLiveCodeCellAndIndefiniteDuration(t *testing.T) {
	html := `<table><tr><th>Code</th><th>Server</th><th>Rewards</th><th>Duration</th></tr><tr><td><b><code>STARRAILGIFT</code></b><br><a>Quick Redeem</a></td><td>All</td><td><span class="item">Stellar Jade ×50</span><span class="item">Credit ×10000</span></td><td>Discovered: April 26, 2023 Valid: (indefinite)</td></tr></table>`
	got, err := parseHTML(html, "test", time.Now(), 1)
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got[0].Code != "STARRAILGIFT" || got[0].Status != domain.StatusActive || got[0].Region != "global" {
		t.Fatalf("candidate=%+v", got[0])
	}
}
