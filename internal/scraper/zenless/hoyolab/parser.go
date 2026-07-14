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
	"cd6682dd2d871dc93dfa28c3f281d527_6175554878133394960": "Dennies",
	"8609070fe148c0e0e367cda25fdae632_208324374592932270":  "Polychrome",
	"6ef3e419022c871257a936b1857ac9d1_411767156105350865":  "W-Engine Energy Module",
	"86e1f7a5ff283d527bbc019475847174_5751095862610622324": "Senior Investigator Logs",
}

func parse(payload apiResponse, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	var candidates []domain.CodeCandidate
	for _, module := range payload.Data.Modules {
		if module.ExchangeGroup == nil {
			continue
		}
		for _, bonus := range module.ExchangeGroup.Bonuses {
			code := strings.TrimSpace(bonus.ExchangeCode)
			if bonus.CodeStatus != "ON" || code == "" || isChina(bonus.Region) || isChina(bonus.Server) {
				continue
			}
			candidates = append(candidates, domain.CodeCandidate{
				GameSlug:   "zenless",
				Code:       code,
				Rewards:    rewards(bonus.IconBonuses),
				Region:     "global",
				Status:     domain.StatusActive,
				SourceID:   "zenless-hoyolab",
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
		name := itemNames[hash]
		if name == "" {
			name = "Unknown reward"
			if hash != "" {
				name += " (" + hash + ")"
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

func iconHash(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	filename := path.Base(parsed.Path)
	return strings.TrimSuffix(filename, path.Ext(filename))
}

func isChina(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "cn" || value == "china" || strings.Contains(value, "china")
}
