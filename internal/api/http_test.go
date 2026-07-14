package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"opengachacodes/internal/domain"
)

type fakeStore struct {
	games      []domain.Game
	codes      []domain.Code
	exists     bool
	err        error
	queriedNow time.Time
}

func (s *fakeStore) ListGames(context.Context) ([]domain.Game, error) { return s.games, s.err }
func (s *fakeStore) GameExists(context.Context, string) (bool, error) { return s.exists, s.err }
func (s *fakeStore) ListActiveCodes(_ context.Context, _ string, now time.Time) ([]domain.Code, error) {
	s.queriedNow = now
	return s.codes, s.err
}

func TestGames(t *testing.T) {
	store := &fakeStore{games: []domain.Game{{Slug: "genshin", Name: "Genshin Impact"}}}
	response := request(Handler{Store: store}, http.MethodGet, "/games")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"slug":"genshin"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestGameCodes(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{exists: true, codes: []domain.Code{{
		GameSlug: "genshin", Code: "GIFT", CanonicalCode: "GIFT",
		Status: domain.StatusActive, Rewards: []string{"Primogem ×50"},
	}}}
	response := request(Handler{Store: store, Now: func() time.Time { return now }}, http.MethodGet, "/games/genshin/codes")
	body := strings.TrimSpace(response.Body.String())
	if response.Code != http.StatusOK || body != `[{"code":"GIFT","rewards":["Primogem x50"]}]` {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
	if !store.queriedNow.Equal(now) {
		t.Fatalf("query time=%v", store.queriedNow)
	}
}

func TestKnownGameWithoutCodesReturnsArray(t *testing.T) {
	response := request(Handler{Store: &fakeStore{exists: true}}, http.MethodGet, "/games/genshin/codes")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestUnknownGameAndPath(t *testing.T) {
	tests := map[string]string{
		"/games/unknown/codes": `"message":"game not found"`,
		"/game/genshin/code":   `"message":"not found"`,
		"/health":              `"message":"not found"`,
		"/api/v1/games":        `"message":"not found"`,
		"/games/":              `"message":"not found"`,
	}
	for path, message := range tests {
		response := request(Handler{Store: &fakeStore{}}, http.MethodGet, path)
		body := strings.TrimSpace(response.Body.String())
		if response.Code != http.StatusNotFound || !strings.Contains(body, `"statusCode":404`) || !strings.Contains(body, message) {
			t.Errorf("%s status=%d body=%s", path, response.Code, body)
		}
	}
}

func TestWrongMethod(t *testing.T) {
	response := request(Handler{Store: &fakeStore{}}, http.MethodPost, "/games")
	body := strings.TrimSpace(response.Body.String())
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodGet || body != `{"statusCode":405,"error":{"message":"method not allowed"}}` {
		t.Fatalf("status=%d allow=%q body=%s", response.Code, response.Header().Get("Allow"), body)
	}
}

func TestRepositoryError(t *testing.T) {
	response := request(Handler{Store: &fakeStore{err: errors.New("database down")}}, http.MethodGet, "/games")
	body := strings.TrimSpace(response.Body.String())
	if response.Code != http.StatusInternalServerError || body != `{"statusCode":500,"error":{"message":"internal server error"}}` || strings.Contains(body, "database down") {
		t.Fatalf("status=%d body=%q", response.Code, body)
	}
}

func request(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}
