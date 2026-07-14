package hoyolab

import (
	"opengachacodes/internal/domain"
	"testing"
	"time"
)

func TestParseActiveCodesAndRewards(t *testing.T) {
	p := apiResponse{}
	p.Data.Modules = []module{{ExchangeGroup: &exchangeGroup{Bonuses: []bonus{{ExchangeCode: "RAIL1234", CodeStatus: "ON", Region: "global", IconBonuses: []iconBonus{{BonusNum: 50, IconURL: "https://x/unknown_hash.png"}}}, {ExchangeCode: "CN1234", CodeStatus: "ON", Region: "CN"}, {ExchangeCode: "OFF1234", CodeStatus: "OFF", Region: "global"}}}}}
	got, e := parse(p, "test", time.Now())
	if e != nil || len(got) != 1 || got[0].GameSlug != "starrail" || got[0].Status != domain.StatusActive || got[0].Rewards[0] != "Unknown reward (unknown_hash) ×50" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
