package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"opengachacodes/internal/domain"
)

type Store interface {
	ListGames(context.Context) ([]domain.Game, error)
	GameExists(context.Context, string) (bool, error)
	ListActiveCodes(context.Context, string, time.Time) ([]domain.Code, error)
}

type Handler struct {
	Store Store
	Now   func() time.Time
}

type codeResponse struct {
	Code    string   `json:"code"`
	Rewards []string `json:"rewards"`
}

type errorResponse struct {
	StatusCode int         `json:"statusCode"`
	Error      errorDetail `json:"error"`
}

type errorDetail struct {
	Message string `json:"message"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.URL.Path == "/games" {
		h.listGames(w, r)
		return
	}
	if slug, ok := gameCodesSlug(r.URL.Path); ok {
		h.listCodes(w, r, slug)
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

func (h Handler) listGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.Store.ListGames(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if games == nil {
		games = []domain.Game{}
	}
	writeJSON(w, http.StatusOK, games)
}

func (h Handler) listCodes(w http.ResponseWriter, r *http.Request, slug string) {
	exists, err := h.Store.GameExists(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	now := time.Now().UTC()
	if h.Now != nil {
		now = h.Now().UTC()
	}
	codes, err := h.Store.ListActiveCodes(r.Context(), slug, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response := make([]codeResponse, 0, len(codes))
	for _, code := range codes {
		rewards := make([]string, len(code.Rewards))
		for i, reward := range code.Rewards {
			rewards[i] = strings.ReplaceAll(reward, "×", "x")
		}
		response = append(response, codeResponse{Code: code.Code, Rewards: rewards})
	}
	writeJSON(w, http.StatusOK, response)
}

func gameCodesSlug(path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "games" || parts[2] != "codes" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		StatusCode: status,
		Error:      errorDetail{Message: message},
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
