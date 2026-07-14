package catalog

import (
	"fmt"
	"net/http"
	"sort"

	"opengachacodes/internal/domain"
	"opengachacodes/internal/scraper"
	"opengachacodes/internal/scraper/endfield"
	"opengachacodes/internal/scraper/genshin"
	"opengachacodes/internal/scraper/nte"
	starrail "opengachacodes/internal/scraper/starrail"
	"opengachacodes/internal/scraper/wuwa"
	"opengachacodes/internal/scraper/zenless"
)

type Factory func(*http.Client, string) []scraper.Source

type Definition struct {
	Game    domain.Game
	Factory Factory
}

type Catalog struct{ definitions []Definition }

func New() Catalog {
	return Catalog{definitions: []Definition{
		{Game: domain.Game{Slug: "genshin", Name: "Genshin Impact"}, Factory: genshin.Sources},
		{Game: domain.Game{Slug: "starrail", Name: "Honkai: Star Rail"}, Factory: starrail.Sources},
		{Game: domain.Game{Slug: "zenless", Name: "Zenless Zone Zero"}, Factory: zenless.Sources},
		{Game: domain.Game{Slug: "endfield", Name: "Arknights: Endfield"}, Factory: endfield.Sources},
		{Game: domain.Game{Slug: "wuwa", Name: "Wuthering Waves"}, Factory: wuwa.Sources},
		{Game: domain.Game{Slug: "nte", Name: "Neverness to Everness"}, Factory: nte.Sources},
	}}
}

func (c Catalog) Games() []domain.Game {
	games := make([]domain.Game, len(c.definitions))
	for i, definition := range c.definitions {
		games[i] = definition.Game
	}
	return games
}

func (c Catalog) Sources(client *http.Client, userAgent string) []scraper.Source {
	var sources []scraper.Source
	for _, definition := range c.definitions {
		sources = append(sources, definition.Factory(client, userAgent)...)
	}
	return sources
}

func (c Catalog) SourcesFor(game, alias string, client *http.Client, userAgent string) ([]scraper.Source, error) {
	for _, definition := range c.definitions {
		if definition.Game.Slug != game {
			continue
		}
		sources := definition.Factory(client, userAgent)
		if alias == "all" {
			return sources, nil
		}
		wanted := game + "-" + alias
		for _, source := range sources {
			if source.ID() == wanted {
				return []scraper.Source{source}, nil
			}
		}
		if alias != "hoyolab" && alias != "game8" && alias != "fandom" {
			return nil, fmt.Errorf("unsupported source %q", alias)
		}
		return nil, fmt.Errorf("source %q is not configured", alias)
	}
	return nil, fmt.Errorf("unsupported game %q", game)
}

func (c Catalog) Select(game, alias string, sources []scraper.Source) ([]scraper.Source, error) {
	if alias == "all" {
		result := make([]scraper.Source, 0, len(sources))
		for _, source := range sources {
			if source.GameSlug() == game {
				result = append(result, source)
			}
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("unsupported game %q", game)
		}
		return result, nil
	}
	if game == "" {
		return nil, fmt.Errorf("unsupported game %q", game)
	}
	wanted := game + "-" + alias
	for _, source := range sources {
		if source.GameSlug() == game && source.ID() == wanted {
			return []scraper.Source{source}, nil
		}
	}
	if alias != "hoyolab" && alias != "game8" && alias != "fandom" {
		return nil, fmt.Errorf("unsupported source %q", alias)
	}
	return nil, fmt.Errorf("source %q is not configured", alias)
}

func (c Catalog) Definition(slug string) (Definition, bool) {
	for _, d := range c.definitions {
		if d.Game.Slug == slug {
			return d, true
		}
	}
	return Definition{}, false
}

func (c Catalog) Slugs() []string {
	result := make([]string, 0, len(c.definitions))
	for _, d := range c.definitions {
		result = append(result, d.Game.Slug)
	}
	sort.Strings(result)
	return result
}
