package scraper

import (
	"context"

	"opengachacodes/internal/domain"
)

type Source interface {
	ID() string
	GameSlug() string
	URL() string
	Collect(context.Context) ([]domain.CodeCandidate, error)
}
