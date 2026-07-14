package game8

import (
	"testing"
	"time"
)

func TestParseHTMLUsesOnlyActiveTable(t *testing.T) {
	page := `<h3 id="hm_1">All ZZZ Codes</h3><table class="a-table"><tr><th>Code</th><th>Rewards</th></tr><tr><td><input class="a-clipboard__textInput" value="ZENLESSGIFT"></td><td><div class="align">Polychrome x50</div><div class="align">Denny x10000</div></td></tr></table><h3 id="hm_2">Livestream</h3><table class="a-table"><tr><th>Code</th><th>Rewards</th></tr><tr><td><input class="a-clipboard__textInput" value="EXPIRED"></td><td>Denny x1</td></tr></table>`
	candidates, err := parseHTML(page, "test", time.Now())
	if err != nil || len(candidates) != 1 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Code != "ZENLESSGIFT" || candidates[0].GameSlug != "zenless" || candidates[0].Rewards[1] != "Denny ×10,000" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
}
