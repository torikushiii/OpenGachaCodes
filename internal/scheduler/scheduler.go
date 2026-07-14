package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	Location *time.Location
	Run      func(context.Context)
	Now      func() time.Time
	Logger   *slog.Logger
}

func (s Scheduler) Start(ctx context.Context) {
	location := s.Location
	if location == nil {
		location = time.Local
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	logger := s.Logger
	if logger == nil {
		logger = slog.Default()
	}

	for {
		next := NextBoundary(now(), location)
		logger.Info("next collection scheduled", "at", next)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}
		if s.Run != nil {
			s.Run(ctx)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func NextBoundary(now time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}
	local := now.In(location)
	hour := time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, location)
	candidate := hour.Add(30 * time.Minute)
	if !local.Before(candidate) {
		candidate = hour.Add(time.Hour)
	}
	return candidate
}
