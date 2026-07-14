package fandom

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var datePatterns = []string{"January 2, 2006", "Jan 2, 2006", "2006-01-02", "2006/01/02", "02 January 2006", "2 January 2006"}

func parseHTML(html, sourceURL string, observed time.Time, revision int64) ([]domain.CodeCandidate, error) {
	doc, e := goquery.NewDocumentFromReader(strings.NewReader(html))
	if e != nil {
		return nil, fmt.Errorf("parse redemption HTML: %w", e)
	}
	var out []domain.CodeCandidate
	doc.Find("table").Each(func(_ int, t *goquery.Selection) {
		headers := tableHeaders(t)
		cc := exactColumn(headers, "code")
		sc := exactColumn(headers, "server")
		rc := exactColumn(headers, "rewards")
		dc := exactColumn(headers, "duration")
		if cc < 0 || sc < 0 || rc < 0 || dc < 0 {
			return
		}
		t.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() <= dc || cells.Length() <= cc {
				return
			}
			code := codeText(cells.Eq(cc))
			if !looksLikeCode(code) || isChina(cells.Eq(sc).Text()) {
				return
			}
			status, expiry := duration(cells.Eq(dc).Text(), observed)
			out = append(out, domain.CodeCandidate{GameSlug: "starrail", Code: code, Rewards: splitRewards(cells.Eq(rc)), Region: region(cells.Eq(sc).Text()), Status: status, ExpiresAt: expiry, SourceID: "starrail-fandom", SourceURL: sourceURL, Authority: domain.AuthorityCommunity, ObservedAt: observed, RevisionID: revision})
		})
	})
	if len(out) == 0 {
		return nil, fmt.Errorf("no redemption codes found in Fandom HTML")
	}
	return out, nil
}
func tableHeaders(t *goquery.Selection) []string {
	var h []string
	t.Find("tr").First().Find("th,td").Each(func(_ int, c *goquery.Selection) { h = append(h, normalizeHeader(c.Text())) })
	return h
}
func normalizeHeader(v string) string { return strings.ToLower(strings.Join(strings.Fields(v), " ")) }
func exactColumn(h []string, name string) int {
	for i, v := range h {
		if v == name {
			return i
		}
	}
	return -1
}
func splitRewards(c *goquery.Selection) []string {
	var out []string
	c.Find("li,.item").Each(func(_ int, x *goquery.Selection) {
		if v := clean(x.Text()); v != "" {
			out = appendUnique(out, v)
		}
	})
	if len(out) == 0 {
		for _, v := range strings.FieldsFunc(clean(c.Text()), func(r rune) bool { return r == '\n' || r == '；' || r == ';' || r == '|' }) {
			if v = clean(v); v != "" {
				out = appendUnique(out, v)
			}
		}
	}
	return out
}
func appendUnique(a []string, v string) []string {
	for _, x := range a {
		if x == v {
			return a
		}
	}
	return append(a, v)
}

func codeText(cell *goquery.Selection) string {
	if code := clean(cell.Find("code").First().Text()); code != "" {
		return code
	}
	return clean(cell.Text())
}

func duration(raw string, now time.Time) (domain.Status, *time.Time) {
	v := clean(raw)
	if v == "" {
		return domain.StatusUnknown, nil
	}
	lower := strings.ToLower(v)
	if strings.Contains(lower, "permanent") || strings.Contains(lower, "indefinite") || strings.Contains(lower, "never expire") || strings.Contains(lower, "no expiration") {
		return domain.StatusActive, nil
	}
	if strings.Contains(lower, "expired") {
		return domain.StatusExpired, nil
	}
	for _, layout := range datePatterns {
		if d, e := time.ParseInLocation(layout, v, now.Location()); e == nil {
			d = d.UTC()
			if !d.After(now) {
				return domain.StatusExpired, &d
			}
			return domain.StatusActive, &d
		}
	}
	if m := regexp.MustCompile(`(?i)(?:until|through|expires?\s*(?:on|at)?)\s*:?\s*([A-Za-z]+\s+\d{1,2},\s*\d{4}|\d{4}-\d{2}-\d{2})`).FindStringSubmatch(v); len(m) > 1 {
		for _, layout := range datePatterns {
			if d, e := time.ParseInLocation(layout, m[1], now.Location()); e == nil {
				d = d.UTC()
				if !d.After(now) {
					return domain.StatusExpired, &d
				}
				return domain.StatusActive, &d
			}
		}
	}
	return domain.StatusUnknown, nil
}
func region(v string) string {
	v = strings.ToLower(clean(v))
	if v == "global" || strings.Contains(v, "all") || strings.Contains(v, "global") {
		return "global"
	}
	return v
}
func clean(v string) string { return strings.Join(strings.Fields(v), " ") }
func looksLikeCode(v string) bool {
	if len(v) < 4 || len(v) > 64 || strings.ContainsAny(v, " \t\r\n") {
		return false
	}
	for _, r := range v {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}
func isChina(v string) bool {
	v = strings.ToLower(clean(v))
	return v == "cn" || v == "china" || strings.Contains(v, "china")
}
