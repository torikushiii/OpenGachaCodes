package fandom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"opengachacodes/internal/domain"
)

const (
	DefaultAPIURL   = "https://genshin-impact.fandom.com/api.php"
	DefaultPageURL  = "https://genshin-impact.fandom.com/wiki/Promotional_Code"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	APIURL    string
	UserAgent string
	Now       func() time.Time
}

type apiResponse struct {
	Parse struct {
		Text       string `json:"text"`
		RevisionID int64  `json:"revid"`
	} `json:"parse"`
	Error *struct {
		Code string `json:"code"`
		Info string `json:"info"`
	} `json:"error,omitempty"`
}

func (s *Source) ID() string       { return "genshin-fandom" }
func (s *Source) GameSlug() string { return "genshin" }
func (s *Source) URL() string      { return DefaultPageURL }

func (s *Source) Collect(ctx context.Context) ([]domain.CodeCandidate, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	apiURL := s.APIURL
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}
	userAgent := strings.TrimSpace(s.UserAgent)
	if userAgent == "" {
		return nil, fmt.Errorf("user agent is required")
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}
	q := u.Query()
	q.Set("action", "parse")
	q.Set("page", "Promotional_Code")
	q.Set("prop", "text|revid")
	q.Set("format", "json")
	q.Set("formatversion", "2")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Fandom API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Fandom API returned %s", resp.Status)
	}

	var payload apiResponse
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize+1))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode Fandom API response: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("Fandom API error %s: %s", payload.Error.Code, payload.Error.Info)
	}
	if strings.TrimSpace(payload.Parse.Text) == "" {
		return nil, fmt.Errorf("Fandom API response has no parsed HTML")
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	candidates, err := parseHTML(payload.Parse.Text, now, payload.Parse.RevisionID)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}
