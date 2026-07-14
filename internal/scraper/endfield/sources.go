package endfield

import (
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/endfield/game8"
	"opengachacodes/internal/scraper/endfield/mobalytics"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&game8.Source{Client: client, UserAgent: userAgent},
		&mobalytics.Source{Client: client, UserAgent: userAgent},
	}
}
