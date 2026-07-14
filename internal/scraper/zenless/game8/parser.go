package game8

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
)

func parseHTML(page, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse Game8 HTML: %w", err)
	}
	heading := doc.Find("h3#hm_1").First()
	if heading.Length() == 0 {
		return nil, fmt.Errorf("Game8 active code section not found")
	}
	table := heading.NextUntil("h2,h3").Filter("table.a-table").First()
	if table.Length() == 0 {
		return nil, fmt.Errorf("Game8 active code table not found")
	}

	headers := tableHeaders(table)
	codeColumn := findColumn(headers, "redeem code", "code")
	rewardColumn := findColumn(headers, "reward")
	if codeColumn < 0 || rewardColumn < 0 {
		return nil, fmt.Errorf("Game8 active table is missing code or reward headers")
	}

	var candidates []domain.CodeCandidate
	table.Find("tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() <= codeColumn || cells.Length() <= rewardColumn {
			return
		}
		code, ok := cells.Eq(codeColumn).Find("input.a-clipboard__textInput").First().Attr("value")
		code = strings.TrimSpace(code)
		if !ok || !looksLikeCode(code) {
			return
		}
		candidates = append(candidates, domain.CodeCandidate{
			GameSlug:   "zenless",
			Code:       code,
			Rewards:    parseRewards(cells.Eq(rewardColumn)),
			Region:     "global",
			Status:     domain.StatusActive,
			SourceID:   "zenless-game8",
			SourceURL:  sourceURL,
			Authority:  domain.AuthorityCommunity,
			ObservedAt: observedAt,
		})
	})
	if len(candidates) == 0 {
		return nil, fmt.Errorf("Game8 active table contains no valid codes")
	}
	return candidates, nil
}

func tableHeaders(table *goquery.Selection) []string {
	var headers []string
	table.Find("tr").First().Find("th,td").Each(func(_ int, cell *goquery.Selection) {
		headers = append(headers, strings.ToLower(cleanText(cell.Text())))
	})
	return headers
}

func findColumn(headers []string, terms ...string) int {
	for i, header := range headers {
		for _, term := range terms {
			if strings.Contains(header, term) {
				return i
			}
		}
	}
	return -1
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
	if len(rewards) == 0 {
		if value := normalizeReward(cleanText(cell.Text())); value != "" {
			rewards = append(rewards, value)
		}
	}
	return rewards
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
