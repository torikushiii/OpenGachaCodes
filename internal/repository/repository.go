package repository

import (
	"context"
	"time"

	"opengachacodes/internal/domain"
)

type Store interface {
	EnsureGames(context.Context, []domain.Game) error
	ListGames(context.Context) ([]domain.Game, error)
	GameExists(context.Context, string) (bool, error)
	UpsertCodes(context.Context, []domain.Code) error
	ListActiveCodes(context.Context, string, time.Time) ([]domain.Code, error)
}
