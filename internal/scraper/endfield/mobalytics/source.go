package mobalytics

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
	DefaultPageURL  = "https://mobalytics.gg/arknights-endfield/guides/redemption-codes"
	DefaultFetchURL = "https://r.jina.ai/https://mobalytics.gg/arknights-endfield/guides/redemption-codes"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	PageURL   string
	FetchURL  string
	UserAgent string
	Now       func() time.Time
}

func (s *Source) ID() string       { return "endfield-mobalytics" }
func (s *Source) GameSlug() string { return "endfield" }
func (s *Source) URL() string {
	if strings.TrimSpace(s.PageURL) != "" {
		return s.PageURL
	}
	return DefaultPageURL
}

func (s *Source) fetchURL() string {
	if strings.TrimSpace(s.FetchURL) != "" {
		return s.FetchURL
	}
	return DefaultFetchURL
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

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.fetchURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Mobalytics request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request Mobalytics page: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("Mobalytics returned %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Mobalytics response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("Mobalytics response exceeds %d bytes", maxResponseSize)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return parseMarkdown(string(body), s.URL(), now)
}
