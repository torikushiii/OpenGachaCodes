package hoyolab

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"opengachacodes/internal/domain"
)

var itemNames = map[string]string{
	"150a941de99e21fc96dce97cde2dae22_1631694835879620915": "Primogem",
	"46de1e881b5dff638969aed85850e388_7373589751062039567": "Hero's Wit",
	"503abf5f2f2c8b2013dde0f2197fc9ac_3214074117670348863": "Mora",
	"d3eb1267f27bead29907cb279d4365ab_4473305467748929436": "Mystic Enhancement Ore",
}

func parse(payload apiResponse, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	var candidates []domain.CodeCandidate
	for _, module := range payload.Data.Modules {
		if module.ExchangeGroup == nil {
			continue
		}
		for _, bonus := range module.ExchangeGroup.Bonuses {
			code := strings.TrimSpace(bonus.ExchangeCode)
			if bonus.CodeStatus != "ON" || code == "" || strings.EqualFold(code, "YUANSHEN") {
				continue
			}
			if isChina(bonus.Region) || isChina(bonus.Server) {
				continue
			}
			candidates = append(candidates, domain.CodeCandidate{
				GameSlug:   "genshin",
				Code:       code,
				Rewards:    rewards(bonus.IconBonuses),
				Region:     "global",
				Status:     domain.StatusActive,
				SourceID:   "genshin-hoyolab",
				SourceURL:  sourceURL,
				Authority:  domain.AuthorityOfficial,
				ObservedAt: observedAt,
			})
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("HoYoLAB response contains no active exchange codes")
	}
	return candidates, nil
}

func rewards(items []iconBonus) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		hash := iconHash(item.IconURL)
		name, known := itemNames[hash]
		if !known {
			if hash == "" {
				name = "Unknown reward"
			} else {
				name = "Unknown reward (" + hash + ")"
			}
		}
		reward := fmt.Sprintf("%s ×%d", name, item.BonusNum)
		if !seen[reward] {
			result = append(result, reward)
			seen[reward] = true
		}
	}
	return result
}

func iconHash(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	filename := path.Base(parsed.Path)
	extension := path.Ext(filename)
	return strings.TrimSuffix(filename, extension)
}

func isChina(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "cn" || value == "china" || strings.Contains(value, "china")
}
