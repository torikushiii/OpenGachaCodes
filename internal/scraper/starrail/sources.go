package starrail

import (
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/starrail/fandom"
	"opengachacodes/internal/scraper/starrail/game8"
	"opengachacodes/internal/scraper/starrail/hoyolab"
	"opengachacodes/internal/scraper/starrail/sportskeeda"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&hoyolab.Source{Client: client, UserAgent: userAgent},
		&game8.Source{Client: client, UserAgent: userAgent},
		&fandom.Source{Client: client, UserAgent: userAgent},
		&sportskeeda.Source{Client: client, UserAgent: userAgent},
	}
}
