package sportskeeda

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
)

var (
	rewardXQuantity = regexp.MustCompile(`^(.+?)\s+[xX]([\d,]+)$`)
	rewardQuantityX = regexp.MustCompile(`^(.+?)\s+([\d,]+)[xX]$`)
	rewardLeading   = regexp.MustCompile(`^([\d,]+)\s+(.+)$`)
)

func parseHTML(page, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse Sportskeeda HTML: %w", err)
	}

	var heading *goquery.Selection
	doc.Find("h2").EachWithBreak(func(_ int, candidate *goquery.Selection) bool {
		text := strings.ToLower(cleanText(candidate.Text()))
		if strings.Contains(text, "active") && strings.Contains(text, "honkai star rail") && strings.Contains(text, "redeem codes") {
			heading = candidate
			return false
		}
		return true
	})
	if heading == nil {
		return nil, fmt.Errorf("Sportskeeda active code section not found")
	}

	section := heading.NextUntil("h2")
	list := section.Filter("ul").First()
	if list.Length() == 0 {
		list = section.Find("ul").First()
	}
	if list.Length() == 0 {
		return nil, fmt.Errorf("Sportskeeda active code list not found")
	}

	var candidates []domain.CodeCandidate
	list.Find("li").Each(func(_ int, item *goquery.Selection) {
		strong := item.Find("strong").First()
		if strong.Length() == 0 {
			return
		}
		label := cleanText(strong.Text())
		code := strings.TrimSuffix(label, ":")
		code = strings.TrimSpace(code)
		if !looksLikeCode(code) {
			return
		}

		text := cleanText(item.Text())
		rewardText := strings.TrimSpace(strings.TrimPrefix(text, label))
		rewardText = strings.TrimSpace(strings.TrimPrefix(rewardText, ":"))
		candidates = append(candidates, domain.CodeCandidate{
			GameSlug:   "starrail",
			Code:       code,
			Rewards:    parseRewards(rewardText),
			Region:     "global",
			Status:     domain.StatusActive,
			SourceID:   "starrail-sportskeeda",
			SourceURL:  sourceURL,
			Authority:  domain.AuthorityCommunity,
			ObservedAt: observedAt,
		})
	})
	if len(candidates) == 0 {
		return nil, fmt.Errorf("Sportskeeda active code list contains no valid codes")
	}
	return candidates, nil
}

func parseRewards(value string) []string {
	parts := splitRewards(value)
	result := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		reward := normalizeReward(cleanText(part))
		if reward != "" && !seen[reward] {
			result = append(result, reward)
			seen[reward] = true
		}
	}
	return result
}

func splitRewards(value string) []string {
	var result []string
	start := 0
	for i, r := range value {
		if r != ',' {
			continue
		}
		after := value[i+1:]
		rest := strings.TrimLeftFunc(after, unicode.IsSpace)
		if rest == "" {
			continue
		}
		if after == rest && !unicode.IsLetter([]rune(rest)[0]) {
			continue
		}
		result = append(result, value[start:i])
		start = i + 1
	}
	return append(result, value[start:])
}

func normalizeReward(value string) string {
	if match := rewardXQuantity.FindStringSubmatch(value); match != nil {
		return cleanText(match[1]) + " ×" + match[2]
	}
	if match := rewardQuantityX.FindStringSubmatch(value); match != nil {
		return cleanText(match[1]) + " ×" + match[2]
	}
	if match := rewardLeading.FindStringSubmatch(value); match != nil {
		return cleanText(match[2]) + " ×" + match[1]
	}
	return value
}

func cleanText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func looksLikeCode(code string) bool {
	if len(code) < 4 || len(code) > 64 || strings.ContainsAny(code, " \t\r\n") {
		return false
	}
	for _, r := range code {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}
