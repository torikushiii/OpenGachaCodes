package mongodb

import (
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

func TestMergeCodePreservesOfficialStatusAcrossCommunityOnlyRun(t *testing.T) {
	first := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	existing := domain.Code{
		Status:  domain.StatusActive,
		Rewards: []string{"Primogem ×60"},
		Sources: []domain.SourceAttribution{{
			SourceID: "genshin-hoyolab", Authority: domain.AuthorityOfficial,
			Status: domain.StatusActive, FirstSeen: first, LastSeen: first,
		}},
	}
	incoming := domain.Code{
		Status:  domain.StatusExpired,
		Rewards: []string{"Mora ×10,000"},
		Sources: []domain.SourceAttribution{{
			SourceID: "genshin-fandom", Authority: domain.AuthorityCommunity,
			Status: domain.StatusExpired, FirstSeen: later, LastSeen: later,
		}},
	}

	merged := mergeCode(existing, incoming)
	if merged.Status != domain.StatusActive {
		t.Fatalf("status=%q, want active", merged.Status)
	}
	if len(merged.Sources) != 2 || len(merged.Rewards) != 2 {
		t.Fatalf("merged=%#v", merged)
	}
}

func TestMergeRewardsReplacesPreviouslyCombinedValue(t *testing.T) {
	got := mergeRewards(
		[]string{"Primogem ×60 Adventurer's Experience ×5"},
		[]string{"Primogem ×60", "Adventurer's Experience ×5"},
	)
	if len(got) != 2 || got[0] != "Adventurer's Experience ×5" || got[1] != "Primogem ×60" {
		t.Fatalf("rewards=%#v", got)
	}
}

func TestMergeCodeUsesLatestOfficialObservation(t *testing.T) {
	first := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	existing := domain.Code{Sources: []domain.SourceAttribution{{
		SourceID: "genshin-hoyolab", Authority: domain.AuthorityOfficial,
		Status: domain.StatusActive, FirstSeen: first, LastSeen: first,
	}}}
	incoming := domain.Code{Sources: []domain.SourceAttribution{{
		SourceID: "genshin-hoyolab", Authority: domain.AuthorityOfficial,
		Status: domain.StatusExpired, FirstSeen: first.Add(time.Hour), LastSeen: first.Add(time.Hour),
	}}}
	if got := mergeCode(existing, incoming).Status; got != domain.StatusExpired {
		t.Fatalf("status=%q, want expired", got)
	}
}
