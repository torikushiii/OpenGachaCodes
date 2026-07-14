package wuwa

import (
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/wuwa/fandom"
	"opengachacodes/internal/scraper/wuwa/game8"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&game8.Source{Client: client, UserAgent: userAgent},
		&fandom.Source{Client: client, UserAgent: userAgent},
	}
}
