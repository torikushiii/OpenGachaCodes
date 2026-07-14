package genshin

import (
	"net/http"
	"testing"
)

func TestSourcesAndSelection(t *testing.T) {
	sources := Sources(http.DefaultClient, "agent")
	want := []string{"genshin-hoyolab", "genshin-game8", "genshin-fandom"}
	if len(sources) != len(want) {
		t.Fatalf("got %d sources", len(sources))
	}
	for i, id := range want {
		if sources[i].ID() != id {
			t.Errorf("source[%d]=%q, want %q", i, sources[i].ID(), id)
		}
	}
	selected, err := Select(sources, "game8")
	if err != nil || len(selected) != 1 || selected[0].ID() != "genshin-game8" {
		t.Fatalf("selected=%#v err=%v", selected, err)
	}
}
