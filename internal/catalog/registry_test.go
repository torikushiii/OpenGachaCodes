package catalog

import (
	"net/http"
	"testing"
)

func TestCatalogRegistersGamesAndSources(t *testing.T) {
	c := New()
	games := c.Games()
	if len(games) != 6 {
		t.Fatalf("games=%d", len(games))
	}
	if games[0].Slug != "genshin" || games[0].Name != "Genshin Impact" || games[1].Slug != "starrail" || games[1].Name != "Honkai: Star Rail" || games[2].Slug != "zenless" || games[2].Name != "Zenless Zone Zero" || games[3].Slug != "endfield" || games[3].Name != "Arknights: Endfield" || games[4].Slug != "wuwa" || games[4].Name != "Wuthering Waves" || games[5].Slug != "nte" || games[5].Name != "Neverness to Everness" {
		t.Fatalf("games=%+v", games)
	}
	sources := c.Sources(http.DefaultClient, "test-agent")
	if len(sources) != 15 {
		t.Fatalf("sources=%d", len(sources))
	}
	selected, err := c.Select("starrail", "sportskeeda", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "starrail-sportskeeda" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	selected, err = c.Select("starrail", "game8", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "starrail-game8" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	selected, err = c.Select("zenless", "fandom", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "zenless-fandom" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	selected, err = c.Select("endfield", "mobalytics", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "endfield-mobalytics" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	selected, err = c.Select("wuwa", "fandom", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "wuwa-fandom" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	selected, err = c.Select("nte", "game8", sources)
	if err != nil || len(selected) != 1 || selected[0].ID() != "nte-game8" {
		t.Fatalf("selected=%v err=%v", selected, err)
	}
	if _, err := c.Select("bad", "all", sources); err == nil {
		t.Fatal("expected invalid game error")
	}
	if _, err := c.Select("starrail", "bad", sources); err == nil {
		t.Fatal("expected invalid source error")
	}
}
