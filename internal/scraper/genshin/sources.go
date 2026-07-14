package genshin

import (
	"fmt"
	"net/http"

	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/genshin/fandom"
	"opengachacodes/internal/scraper/genshin/game8"
	"opengachacodes/internal/scraper/genshin/hoyolab"
)

func Sources(client *http.Client, userAgent string) []scraper.Source {
	return []scraper.Source{
		&hoyolab.Source{Client: client, UserAgent: userAgent},
		&game8.Source{Client: client, UserAgent: userAgent},
		&fandom.Source{Client: client, UserAgent: userAgent},
	}
}

func Select(sources []scraper.Source, name string) ([]scraper.Source, error) {
	if name == "all" {
		return sources, nil
	}
	ids := map[string]string{
		"hoyolab": "genshin-hoyolab",
		"game8":   "genshin-game8",
		"fandom":  "genshin-fandom",
	}
	wanted, ok := ids[name]
	if !ok {
		return nil, fmt.Errorf("unsupported source %q", name)
	}
	for _, source := range sources {
		if source.ID() == wanted {
			return []scraper.Source{source}, nil
		}
	}
	return nil, fmt.Errorf("source %q is not configured", name)
}
