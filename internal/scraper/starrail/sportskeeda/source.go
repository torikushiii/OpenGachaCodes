package sportskeeda

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
	DefaultPageURL   = "https://www.sportskeeda.com/esports/honkai-star-rail-hsr-4-0-redeem-codes"
	maxResponseSize  = 8 << 20
	browserUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
)

type Source struct {
	Client    *http.Client
	PageURL   string
	UserAgent string
	Now       func() time.Time
}

func (s *Source) ID() string       { return "starrail-sportskeeda" }
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
	userAgent := strings.TrimSpace(s.UserAgent)
	if userAgent == "" {
		return nil, fmt.Errorf("user agent is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Sportskeeda request: %w", err)
	}
	req.Header.Set("User-Agent", browserUserAgent+" "+userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Sportskeeda page: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("Sportskeeda returned %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Sportskeeda response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("Sportskeeda response exceeds %d bytes", maxResponseSize)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return parseHTML(string(body), s.URL(), now)
}
