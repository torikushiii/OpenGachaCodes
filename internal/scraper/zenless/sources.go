package zenless

import (
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/zenless/fandom"
	"opengachacodes/internal/scraper/zenless/game8"
	"opengachacodes/internal/scraper/zenless/hoyolab"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&hoyolab.Source{Client: client, UserAgent: userAgent},
		&game8.Source{Client: client, UserAgent: userAgent},
		&fandom.Source{Client: client, UserAgent: userAgent},
	}
}
