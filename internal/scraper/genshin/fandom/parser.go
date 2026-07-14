package fandom

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
)

func parseHTML(html string, observedAt time.Time, revisionID int64) ([]domain.CodeCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse promotional code HTML: %w", err)
	}

	var candidates []domain.CodeCandidate
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		headers := tableHeaders(table)
		codeColumn := findColumn(headers, "code")
		if codeColumn < 0 {
			return
		}
		rewardColumn := findColumn(headers, "reward")
		expiryColumn := findColumn(headers, "expire", "valid", "duration")
		serverColumn := findColumn(headers, "server", "region")

		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() <= codeColumn {
				return
			}
			code := cleanText(cells.Eq(codeColumn).Text())
			if !looksLikeCode(code) {
				return
			}
			if serverColumn >= 0 && cells.Length() > serverColumn {
				server := strings.ToLower(cleanText(cells.Eq(serverColumn).Text()))
				if strings.Contains(server, "china") {
					return
				}
			}

			candidate := domain.CodeCandidate{
				GameSlug:   "genshin",
				Code:       code,
				Region:     "global",
				Status:     domain.StatusActive,
				SourceID:   "genshin-fandom",
				SourceURL:  DefaultPageURL,
				Authority:  domain.AuthorityCommunity,
				ObservedAt: observedAt,
				RevisionID: revisionID,
			}
			if rewardColumn >= 0 && cells.Length() > rewardColumn {
				candidate.Rewards = splitRewards(cells.Eq(rewardColumn))
			}
			if expiryColumn >= 0 && cells.Length() > expiryColumn {
				status := strings.ToLower(cleanText(cells.Eq(expiryColumn).Text()))
				if strings.Contains(status, "expired") {
					candidate.Status = domain.StatusExpired
				}
			}
			candidates = append(candidates, candidate)
		})
	})
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no promotional codes found in Fandom HTML")
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

func splitRewards(cell *goquery.Selection) []string {
	var rewards []string
	cell.Find(".item").Each(func(_ int, item *goquery.Selection) {
		if reward := cleanText(item.Text()); reward != "" {
			rewards = append(rewards, reward)
		}
	})
	if len(rewards) == 0 {
		cell.Find("li").Each(func(_ int, item *goquery.Selection) {
			if reward := cleanText(item.Text()); reward != "" {
				rewards = append(rewards, reward)
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
