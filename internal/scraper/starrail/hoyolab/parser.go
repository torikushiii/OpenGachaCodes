package hoyolab

import (
	"fmt"
	"net/url"
	"opengachacodes/internal/domain"
	"path"
	"strings"
	"time"
)

var itemNames = map[string]string{
	"5a75e8e6d0c8d8be4b8d9f5c4b7c1e3f_1631694835879620915": "Stellar Jade",
	"46de1e881b5dff638969aed85850e388_7373589751062039567": "Traveler's Guide",
	"d3eb1267f27bead29907cb279d4365ab_4473305467748929436": "Refined Aether",
	"503abf5f2f2c8b2013dde0f2197fc9ac_3214074117670348863": "Credit",
}

func parse(payload apiResponse, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	var out []domain.CodeCandidate
	for _, m := range payload.Data.Modules {
		if m.ExchangeGroup == nil {
			continue
		}
		for _, b := range m.ExchangeGroup.Bonuses {
			code := strings.TrimSpace(b.ExchangeCode)
			if b.CodeStatus != "ON" || code == "" || isChina(b.Region) || isChina(b.Server) {
				continue
			}
			out = append(out, domain.CodeCandidate{GameSlug: "starrail", Code: code, Rewards: rewards(b.IconBonuses), Region: "global", Status: domain.StatusActive, SourceID: "starrail-hoyolab", SourceURL: sourceURL, Authority: domain.AuthorityOfficial, ObservedAt: observedAt})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("HoYoLAB response contains no active exchange codes")
	}
	return out, nil
}
func rewards(items []iconBonus) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		hash := iconHash(item.IconURL)
		name := itemNames[hash]
		if name == "" {
			name = "Unknown reward"
			if hash != "" {
				name += " (" + hash + ")"
			}
		}
		r := fmt.Sprintf("%s ×%d", name, item.BonusNum)
		if !seen[r] {
			out = append(out, r)
			seen[r] = true
		}
	}
	return out
}
func iconHash(raw string) string {
	u, e := url.Parse(raw)
	if e != nil {
		return ""
	}
	f := path.Base(u.Path)
	return strings.TrimSuffix(f, path.Ext(f))
}
func isChina(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "cn" || v == "china" || strings.Contains(v, "china")
}
