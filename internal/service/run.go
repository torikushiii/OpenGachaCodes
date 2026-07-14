package service

import (
	"context"
	"fmt"
	"time"

	"opengachacodes/internal/repository"
	"opengachacodes/internal/scraper"
)

type Runner struct {
	Store         repository.Store
	Sources       []scraper.Source
	SourceTimeout time.Duration
}

func (r Runner) Run(ctx context.Context) CollectionResult {
	result := CollectWithTimeout(ctx, r.Sources, r.SourceTimeout)
	if len(result.Codes) == 0 {
		return result
	}
	if err := r.Store.UpsertCodes(ctx, result.Codes); err != nil {
		result.Errors = append(result.Errors, SourceError{SourceID: "mongodb", Err: fmt.Errorf("persist codes: %w", err)})
	}
	return result
}
