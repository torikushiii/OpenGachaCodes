package game8

import (
	"strings"
	"testing"
	"time"
)

func TestParseHTMLUsesOnlyActiveTableAndFiltersChina(t *testing.T) {
	html := `<h3 id="hm_1">Active</h3><table class="a-table"><tr><th>Redeem Code</th><th>Reward</th><th>Server</th></tr><tr><td><input class="a-clipboard__textInput" value="STAR1234"></td><td><div class="align">Stellar Jade x 50</div></td><td>Global</td></tr><tr><td><input class="a-clipboard__textInput" value="CN1234"></td><td>Credit x 1</td><td>China</td></tr></table><h3>Expired</h3><table class="a-table"><tr><th>Code</th><th>Reward</th></tr><tr><td><input class="a-clipboard__textInput" value="DECOY123"></td><td>Credit x 1</td></tr></table>`
	got, e := parseHTML(html, "test", time.Now())
	if e != nil || len(got) != 1 || got[0].Code != "STAR1234" || !strings.Contains(got[0].Rewards[0], "×50") {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
