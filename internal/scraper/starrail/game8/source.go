package game8

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"opengachacodes/internal/domain"
	"strings"
	"time"
)

const (
	DefaultPageURL  = "https://game8.co/games/Honkai-Star-Rail/archives/410296"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	PageURL   string
	UserAgent string
	Now       func() time.Time
}

func (s *Source) ID() string       { return "starrail-game8" }
func (s *Source) GameSlug() string { return "starrail" }
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
	ua := strings.TrimSpace(s.UserAgent)
	if ua == "" {
		return nil, fmt.Errorf("user agent is required")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, s.URL(), nil)
	if e != nil {
		return nil, fmt.Errorf("create Game8 request: %w", e)
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html")
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("request Game8 page: %w", e)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Game8 returned %s", resp.Status)
	}
	body, e := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if e != nil {
		return nil, fmt.Errorf("read Game8 response: %w", e)
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
