package mobalytics

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"opengachacodes/internal/domain"
)

var (
	codePrefix       = regexp.MustCompile(`^([A-Za-z0-9_-]{4,64})\b`)
	rewardXQuantity  = regexp.MustCompile(`^(.+?)\s+[xX]([\d,]+)$`)
	rewardLeadingQty = regexp.MustCompile(`^([\d,]+)\s+(.+)$`)
)

func parseMarkdown(markdown, sourceURL string, observedAt time.Time) ([]domain.CodeCandidate, error) {
	inActiveSection := false
	var candidates []domain.CodeCandidate
	for _, line := range strings.Split(markdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			heading := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			if inActiveSection && !strings.Contains(heading, "active promo codes") {
				break
			}
			inActiveSection = strings.Contains(heading, "endfield active promo codes")
			continue
		}
		if !inActiveSection || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		if len(cells) < 4 {
			continue
		}
		codeCell := cleanMarkdown(cells[1])
		lowerCodeCell := strings.ToLower(codeCell)
		if strings.Contains(lowerCodeCell, "latest codes") || strings.Contains(lowerCodeCell, "expired") {
			continue
		}
		match := codePrefix.FindStringSubmatch(codeCell)
		if match == nil || !looksLikeCode(match[1]) {
			continue
		}
		candidates = append(candidates, domain.CodeCandidate{
			GameSlug:   "endfield",
			Code:       match[1],
			Rewards:    parseRewards(cells[2]),
			Region:     "global",
			Status:     domain.StatusActive,
			SourceID:   "endfield-mobalytics",
			SourceURL:  sourceURL,
			Authority:  domain.AuthorityCommunity,
			ObservedAt: observedAt,
		})
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("Mobalytics active code table contains no valid codes")
	}
	return candidates, nil
}

func parseRewards(value string) []string {
	value = strings.ReplaceAll(value, "**", "")
	var rewards []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(value, "*") {
		part = cleanMarkdown(part)
		if index := strings.Index(strings.ToUpper(part), "VALID UNTIL"); index >= 0 {
			part = strings.TrimSpace(part[:index])
		}
		if part == "" {
			continue
		}
		reward := normalizeReward(part)
		if reward != "" && !seen[reward] {
			rewards = append(rewards, reward)
			seen[reward] = true
		}
	}
	return rewards
}

func normalizeReward(value string) string {
	if match := rewardXQuantity.FindStringSubmatch(value); match != nil {
		quantity, err := strconv.ParseUint(strings.ReplaceAll(match[2], ",", ""), 10, 64)
		if err == nil {
			return normalizeRewardName(cleanMarkdown(match[1])) + " ×" + formatQuantity(quantity)
		}
	}
	if match := rewardLeadingQty.FindStringSubmatch(value); match != nil {
		quantity, err := strconv.ParseUint(strings.ReplaceAll(match[1], ",", ""), 10, 64)
		if err == nil {
			return normalizeRewardName(cleanMarkdown(match[2])) + " ×" + formatQuantity(quantity)
		}
	}
	return value
}

func normalizeRewardName(name string) string {
	switch name {
	case "T-Credits", "T Credits":
		return "T-Creds"
	default:
		return name
	}
}

func formatQuantity(quantity uint64) string {
	digits := strconv.FormatUint(quantity, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
}

func cleanMarkdown(value string) string {
	value = strings.NewReplacer("**", "", "__", "", "`", "", "#", "").Replace(value)
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
