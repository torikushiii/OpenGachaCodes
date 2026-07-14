package game8

import (
	"testing"
	"time"
)

func TestParseHTMLUsesOnlyActiveTable(t *testing.T) {
	page := `<h3 id="hm_1">Livestream</h3><table class="a-table"><tr><th>Code</th><th>Reward</th></tr><tr><td><input class="a-clipboard__textInput" value="STREAMCODE"></td><td>T-Creds x1</td></tr></table><h3 id="hm_2">All Active Codes</h3><table class="a-table"><tr><th>Redeem Codes</th><th>Reward</th></tr><tr><td><input class="a-clipboard__textInput" value="ENDFIELDGIFT"></td><td><div class="align">T-Creds x13000</div><div class="align">Advanced Combat Record x2</div></td></tr></table><h3 id="hm_3">Expired</h3><table class="a-table"><tr><th>Code</th><th>Reward</th></tr><tr><td>OLDCODE</td><td>T-Creds x1</td></tr></table>`
	candidates, err := parseHTML(page, "test", time.Now())
	if err != nil || len(candidates) != 1 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Code != "ENDFIELDGIFT" || candidates[0].GameSlug != "endfield" || candidates[0].Rewards[0] != "T-Creds ×13,000" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
}
