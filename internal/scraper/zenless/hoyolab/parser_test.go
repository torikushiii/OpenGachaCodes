package hoyolab

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestParseActiveCodesAndRewards(t *testing.T) {
	payload := apiResponse{}
	payload.Data.Modules = []module{{ExchangeGroup: &exchangeGroup{Bonuses: []bonus{
		{ExchangeCode: "ZZZCODE", CodeStatus: "ON", IconBonuses: []iconBonus{{BonusNum: 60, IconURL: "https://example.com/8609070fe148c0e0e367cda25fdae632_208324374592932270.png"}}},
		{ExchangeCode: "CNCODE", CodeStatus: "ON", Region: "China"},
		{ExchangeCode: "OFFCODE", CodeStatus: "OFF"},
	}}}}
	candidates, err := parse(payload, "test", time.Now())
	if err != nil || len(candidates) != 1 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].GameSlug != "zenless" || candidates[0].Status != domain.StatusActive || candidates[0].Rewards[0] != "Polychrome ×60" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
}
