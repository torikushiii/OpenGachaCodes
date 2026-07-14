package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"opengachacodes/internal/domain"
	"opengachacodes/internal/scraper"
)

type fakeSource struct {
	id         string
	candidates []domain.CodeCandidate
	err        error
}

type blockingSource struct{ id string }

func (s blockingSource) ID() string       { return s.id }
func (s blockingSource) GameSlug() string { return "genshin" }
func (s blockingSource) URL() string      { return "https://example.com/" + s.id }
func (s blockingSource) Collect(ctx context.Context) ([]domain.CodeCandidate, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (s fakeSource) ID() string       { return s.id }
func (s fakeSource) GameSlug() string { return "genshin" }
func (s fakeSource) URL() string      { return "https://example.com/" + s.id }
func (s fakeSource) Collect(context.Context) ([]domain.CodeCandidate, error) {
	return s.candidates, s.err
}

func TestCollectWithTimeoutContinuesAfterSourceTimeout(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	result := CollectWithTimeout(context.Background(), []scraper.Source{
		blockingSource{id: "slow"},
		fakeSource{id: "working", candidates: []domain.CodeCandidate{{
			GameSlug: "genshin", Code: "GIFT", Status: domain.StatusActive, SourceID: "working", ObservedAt: now,
		}}},
	}, 10*time.Millisecond)
	if len(result.Errors) != 1 || result.Errors[0].SourceID != "slow" {
		t.Fatalf("errors=%+v", result.Errors)
	}
	if len(result.Codes) != 1 || result.Codes[0].Code != "GIFT" {
		t.Fatalf("codes=%+v", result.Codes)
	}
}

func TestCollectMergesSourcesAndPreservesPartialSuccess(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	result := Collect(context.Background(), []scraper.Source{
		fakeSource{id: "community", candidates: []domain.CodeCandidate{{
			GameSlug: "genshin", Code: " gift123 ", Rewards: []string{"Mora"}, SourceID: "community", SourceURL: "community-url", Authority: domain.AuthorityCommunity, ObservedAt: now,
		}}},
		fakeSource{id: "official", candidates: []domain.CodeCandidate{{
			GameSlug: "genshin", Code: "GIFT123", Rewards: []string{"Primogems"}, Region: "global", Status: domain.StatusActive, SourceID: "official", SourceURL: "official-url", Authority: domain.AuthorityOfficial, ObservedAt: now,
		}}},
		fakeSource{id: "broken", err: errors.New("unavailable")},
	})
	if len(result.Codes) != 1 || len(result.Errors) != 1 {
		t.Fatalf("codes=%d errors=%d", len(result.Codes), len(result.Errors))
	}
	code := result.Codes[0]
	if code.CanonicalCode != "GIFT123" || len(code.Sources) != 2 || len(code.Rewards) != 2 {
		t.Fatalf("unexpected merged code: %#v", code)
	}
	if code.Status != domain.StatusActive {
		t.Fatalf("status = %q", code.Status)
	}
}
