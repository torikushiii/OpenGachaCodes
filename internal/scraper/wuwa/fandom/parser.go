package fandom

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
	datePatterns = []string{"January 2, 2006", "Jan 2, 2006", "2006-01-02", "2006/01/02", "02 January 2006", "2 January 2006"}
	expiryDate   = regexp.MustCompile(`(?i)(?:until|through|expires?\s*(?:on|at)?)\s*:?\s*([A-Za-z]+\s+\d{1,2},\s*\d{4}|\d{4}-\d{2}-\d{2})`)
)

func parseHTML(page, sourceURL string, observedAt time.Time, revisionID int64) ([]domain.CodeCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse redemption HTML: %w", err)
	}

	var candidates []domain.CodeCandidate
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		headers := tableHeaders(table)
		codeColumn := exactColumn(headers, "code")
		serverColumn := exactColumn(headers, "server")
		rewardColumn := exactColumn(headers, "rewards")
		durationColumn := exactColumn(headers, "duration")
		if codeColumn < 0 || serverColumn < 0 || rewardColumn < 0 || durationColumn < 0 {
			return
		}
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() <= durationColumn {
				return
			}
			code := codeText(cells.Eq(codeColumn))
			if !looksLikeCode(code) || isChina(cells.Eq(serverColumn).Text()) {
				return
			}
			status, expiresAt := duration(cells.Eq(durationColumn).Text(), observedAt)
			candidates = append(candidates, domain.CodeCandidate{
				GameSlug:   "wuwa",
				Code:       code,
				Rewards:    splitRewards(cells.Eq(rewardColumn)),
				Region:     "global",
				Status:     status,
				ExpiresAt:  expiresAt,
				SourceID:   "wuwa-fandom",
				SourceURL:  sourceURL,
				Authority:  domain.AuthorityCommunity,
				ObservedAt: observedAt,
				RevisionID: revisionID,
			})
		})
	})
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no redemption codes found in Fandom HTML")
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

func exactColumn(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}

func codeText(cell *goquery.Selection) string {
	if code := cleanText(cell.Find("code").First().Text()); code != "" {
		return code
	}
	return cleanText(cell.Text())
}

func splitRewards(cell *goquery.Selection) []string {
	var rewards []string
	cell.Find(".card-container").Each(func(_ int, card *goquery.Selection) {
		name := cleanText(card.Find("img").First().AttrOr("alt", ""))
		quantity := cleanText(card.Find(".card-text").First().Text())
		if name == "" {
			return
		}
		if quantity != "" {
			name += " ×" + strings.TrimPrefix(quantity, "x")
		}
		rewards = appendUnique(rewards, name)
	})
	if len(rewards) == 0 {
		cell.Find("li,.item").Each(func(_ int, item *goquery.Selection) {
			if reward := cleanText(item.Text()); reward != "" {
				rewards = appendUnique(rewards, reward)
			}
		})
	}
	if len(rewards) == 0 {
		if reward := cleanText(cell.Text()); reward != "" {
			rewards = append(rewards, reward)
		}
	}
	return rewards
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func duration(raw string, now time.Time) (domain.Status, *time.Time) {
	value := cleanText(raw)
	if value == "" {
		return domain.StatusUnknown, nil
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "expired") {
		return domain.StatusExpired, nil
	}
	if strings.Contains(lower, "permanent") || strings.Contains(lower, "indefinite") || strings.Contains(lower, "never expire") || strings.Contains(lower, "no expiration") || strings.Contains(lower, "valid until: unknown") {
		return domain.StatusActive, nil
	}
	if match := expiryDate.FindStringSubmatch(value); len(match) > 1 {
		for _, layout := range datePatterns {
			if parsed, err := time.ParseInLocation(layout, match[1], now.Location()); err == nil {
				parsed = parsed.UTC()
				if !parsed.After(now) {
					return domain.StatusExpired, &parsed
				}
				return domain.StatusActive, &parsed
			}
		}
	}
	return domain.StatusUnknown, nil
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

func isChina(value string) bool {
	value = strings.ToLower(cleanText(value))
	return value == "cn" || value == "china" || strings.Contains(value, "china")
}
