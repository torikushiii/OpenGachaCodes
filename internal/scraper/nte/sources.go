package nte

import (
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/nte/game8"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&game8.Source{Client: client, UserAgent: userAgent},
	}
}
