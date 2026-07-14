package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"opengachacodes/internal/domain"
	"opengachacodes/internal/scraper"
)

type SourceError struct {
	SourceID string
	Err      error
}

type CollectionResult struct {
	Codes  []domain.Code
	Errors []SourceError
}

func Collect(ctx context.Context, sources []scraper.Source) CollectionResult {
	return CollectWithTimeout(ctx, sources, 0)
}

func CollectWithTimeout(ctx context.Context, sources []scraper.Source, timeout time.Duration) CollectionResult {
	result := CollectionResult{}
	merged := make(map[string]*domain.Code)

	for _, source := range sources {
		sourceCtx := ctx
		cancel := func() {}
		if timeout > 0 {
			sourceCtx, cancel = context.WithTimeout(ctx, timeout)
		}
		candidates, err := source.Collect(sourceCtx)
		cancel()
		if err != nil {
			result.Errors = append(result.Errors, SourceError{SourceID: source.ID(), Err: err})
			continue
		}
		for _, candidate := range candidates {
			mergeCandidate(merged, candidate)
		}
	}

	for _, code := range merged {
		sort.Strings(code.Rewards)
		sort.Slice(code.Sources, func(i, j int) bool { return code.Sources[i].SourceID < code.Sources[j].SourceID })
		result.Codes = append(result.Codes, *code)
	}
	sort.Slice(result.Codes, func(i, j int) bool {
		return result.Codes[i].CanonicalCode < result.Codes[j].CanonicalCode
	})
	return result
}

func mergeCandidate(merged map[string]*domain.Code, candidate domain.CodeCandidate) {
	canonical := strings.ToUpper(strings.TrimSpace(candidate.Code))
	if canonical == "" || candidate.GameSlug == "" {
		return
	}
	region := strings.ToLower(strings.TrimSpace(candidate.Region))
	if region == "" {
		region = "global"
	}
	key := fmt.Sprintf("%s\x00%s\x00%s", candidate.GameSlug, canonical, region)
	code, exists := merged[key]
	if !exists {
		status := candidate.Status
		if status == "" {
			status = domain.StatusUnknown
		}
		code = &domain.Code{
			GameSlug:      candidate.GameSlug,
			Code:          strings.TrimSpace(candidate.Code),
			CanonicalCode: canonical,
			Region:        region,
			Status:        status,
			ExpiresAt:     candidate.ExpiresAt,
		}
		merged[key] = code
	}
	code.Rewards = appendUnique(code.Rewards, candidate.Rewards...)
	mergeAttribution(code, candidate)
	if candidate.Authority == domain.AuthorityOfficial || code.Status == domain.StatusUnknown {
		if candidate.Status != "" {
			code.Status = candidate.Status
		}
		if candidate.ExpiresAt != nil {
			code.ExpiresAt = candidate.ExpiresAt
		}
	}
}

func mergeAttribution(code *domain.Code, candidate domain.CodeCandidate) {
	for i := range code.Sources {
		if code.Sources[i].SourceID != candidate.SourceID {
			continue
		}
		if candidate.ObservedAt.Before(code.Sources[i].FirstSeen) {
			code.Sources[i].FirstSeen = candidate.ObservedAt
		}
		if candidate.ObservedAt.After(code.Sources[i].LastSeen) {
			code.Sources[i].LastSeen = candidate.ObservedAt
			code.Sources[i].RevisionID = candidate.RevisionID
			if candidate.Status != "" {
				code.Sources[i].Status = candidate.Status
			}
			if candidate.ExpiresAt != nil {
				code.Sources[i].ExpiresAt = candidate.ExpiresAt
			}
		}
		return
	}
	code.Sources = append(code.Sources, domain.SourceAttribution{
		SourceID:   candidate.SourceID,
		SourceURL:  candidate.SourceURL,
		Authority:  candidate.Authority,
		Status:     candidate.Status,
		ExpiresAt:  candidate.ExpiresAt,
		FirstSeen:  candidate.ObservedAt,
		LastSeen:   candidate.ObservedAt,
		RevisionID: candidate.RevisionID,
	})
}

func appendUnique(existing []string, values ...string) []string {
	seen := make(map[string]bool, len(existing))
	for _, value := range existing {
		seen[value] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			existing = append(existing, value)
			seen[value] = true
		}
	}
	return existing
}
