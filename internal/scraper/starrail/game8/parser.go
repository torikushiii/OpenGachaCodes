package game8

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"opengachacodes/internal/domain"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func parseHTML(page, url string, observed time.Time) ([]domain.CodeCandidate, error) {
	doc, e := goquery.NewDocumentFromReader(strings.NewReader(page))
	if e != nil {
		return nil, fmt.Errorf("parse Game8 HTML: %w", e)
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
	cc := findColumn(headers, "redeem code", "code")
	rc := findColumn(headers, "reward")
	sc := findColumn(headers, "server", "region")
	if cc < 0 || rc < 0 {
		return nil, fmt.Errorf("Game8 active table is missing code or reward headers")
	}
	var out []domain.CodeCandidate
	table.Find("tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() <= cc || cells.Length() <= rc {
			return
		}
		input, ok := cells.Eq(cc).Find("input.a-clipboard__textInput").First().Attr("value")
		code := strings.TrimSpace(input)
		if !ok || !looksLikeCode(code) {
			return
		}
		if sc >= 0 && cells.Length() > sc && isChina(cells.Eq(sc).Text()) {
			return
		}
		out = append(out, domain.CodeCandidate{GameSlug: "starrail", Code: code, Rewards: parseRewards(cells.Eq(rc)), Region: "global", Status: domain.StatusActive, SourceID: "starrail-game8", SourceURL: url, Authority: domain.AuthorityCommunity, ObservedAt: observed})
	})
	if len(out) == 0 {
		return nil, fmt.Errorf("Game8 active table contains no valid codes")
	}
	return out, nil
}
func tableHeaders(t *goquery.Selection) []string {
	var h []string
	t.Find("tr").First().Find("th,td").Each(func(_ int, c *goquery.Selection) { h = append(h, strings.ToLower(clean(c.Text()))) })
	return h
}
func findColumn(h []string, terms ...string) int {
	for i, v := range h {
		for _, term := range terms {
			if strings.Contains(v, term) {
				return i
			}
		}
	}
	return -1
}
func parseRewards(c *goquery.Selection) []string {
	var out []string
	seen := map[string]bool{}
	c.Find("div.align").Each(func(_ int, r *goquery.Selection) {
		v := normalize(clean(r.Text()))
		if v != "" && !seen[v] {
			out = append(out, v)
			seen[v] = true
		}
	})
	if len(out) == 0 {
		if v := normalize(clean(c.Text())); v != "" {
			out = []string{v}
		}
	}
	return out
}
func normalize(v string) string {
	v = clean(v)
	i := strings.LastIndex(v, " x")
	if i < 1 {
		return v
	}
	n, e := strconv.ParseUint(strings.ReplaceAll(strings.TrimSpace(v[i+2:]), ",", ""), 10, 64)
	if e != nil {
		return v
	}
	return v[:i] + " ×" + format(n)
}
func format(n uint64) string {
	d := strconv.FormatUint(n, 10)
	for i := len(d) - 3; i > 0; i -= 3 {
		d = d[:i] + "," + d[i:]
	}
	return d
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
