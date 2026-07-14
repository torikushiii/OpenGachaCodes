package game8

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
)

var expiryDate = regexp.MustCompile(`(?i)expiry:\s*([A-Za-z]+\s+\d{1,2},\s*\d{4})`)

func parseHTML(page, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse Game8 HTML: %w", err)
	}
	activeHeading := doc.Find("h2#hl_1").First()
	if activeHeading.Length() == 0 {
		return nil, fmt.Errorf("Game8 active code section not found")
	}
	tables := activeHeading.NextUntil("h2").Filter("table.a-table")
	if tables.Length() == 0 {
		return nil, fmt.Errorf("Game8 active code tables not found")
	}

	var candidates []domain.CodeCandidate
	tables.Find("input.a-clipboard__textInput").Each(func(_ int, input *goquery.Selection) {
		code, ok := input.Attr("value")
		code = strings.TrimSpace(code)
		if !ok || !looksLikeCode(code) {
			return
		}
		cell := input.Closest("td")
		status, expiresAt := expiration(cell.Text(), observedAt)
		candidates = append(candidates, domain.CodeCandidate{
			GameSlug:   "wuwa",
			Code:       code,
			Rewards:    parseRewards(cell),
			Region:     "global",
			Status:     status,
			ExpiresAt:  expiresAt,
			SourceID:   "wuwa-game8",
			SourceURL:  sourceURL,
			Authority:  domain.AuthorityCommunity,
			ObservedAt: observedAt,
		})
	})
	if len(candidates) == 0 {
		return nil, fmt.Errorf("Game8 active section contains no valid codes")
	}
	return candidates, nil
}

func parseRewards(cell *goquery.Selection) []string {
	var rewards []string
	seen := make(map[string]bool)
	cell.Find("div.align").Each(func(_ int, reward *goquery.Selection) {
		value := normalizeReward(cleanText(reward.Text()))
		if value != "" && !seen[value] {
			rewards = append(rewards, value)
			seen[value] = true
		}
	})
	return rewards
}

func expiration(value string, now time.Time) (domain.Status, *time.Time) {
	match := expiryDate.FindStringSubmatch(cleanText(value))
	if len(match) < 2 {
		return domain.StatusActive, nil
	}
	parsed, err := time.ParseInLocation("January 2, 2006", match[1], now.Location())
	if err != nil {
		return domain.StatusActive, nil
	}
	parsed = parsed.UTC()
	if !parsed.After(now) {
		return domain.StatusExpired, &parsed
	}
	return domain.StatusActive, &parsed
}

func normalizeReward(value string) string {
	index := strings.LastIndex(value, " x")
	if index < 1 {
		return value
	}
	quantity, err := strconv.ParseUint(strings.ReplaceAll(value[index+2:], ",", ""), 10, 64)
	if err != nil {
		return value
	}
	return value[:index] + " ×" + formatQuantity(quantity)
}

func formatQuantity(quantity uint64) string {
	digits := strconv.FormatUint(quantity, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
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
