package game8

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"opengachacodes/internal/domain"
)

const (
	DefaultPageURL  = "https://game8.co/games/Zenless-Zone-Zero/archives/435683"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	PageURL   string
	UserAgent string
	Now       func() time.Time
}

func (s *Source) ID() string       { return "zenless-game8" }
func (s *Source) GameSlug() string { return "zenless" }
func (s *Source) URL() string {
	if strings.TrimSpace(s.PageURL) != "" {
		return s.PageURL
	}
	return DefaultPageURL
}

func (s *Source) Collect(ctx context.Context) ([]domain.CodeCandidate, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	userAgent := strings.TrimSpace(s.UserAgent)
	if userAgent == "" {
		return nil, fmt.Errorf("user agent is required")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Game8 request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "text/html")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request Game8 page: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("Game8 returned %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Game8 response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("Game8 response exceeds %d bytes", maxResponseSize)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return parseHTML(string(body), s.URL(), now)
}
