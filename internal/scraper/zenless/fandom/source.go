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
	DefaultAPIURL   = "https://zenless-zone-zero.fandom.com/api.php"
	DefaultPageURL  = "https://zenless-zone-zero.fandom.com/wiki/Redemption_Code"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	APIURL    string
	PageURL   string
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

func (s *Source) ID() string       { return "zenless-fandom" }
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
	apiURL := strings.TrimSpace(s.APIURL)
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}
	userAgent := strings.TrimSpace(s.UserAgent)
	if userAgent == "" {
		return nil, fmt.Errorf("user agent is required")
	}

	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}
	query := parsedURL.Query()
	query.Set("action", "parse")
	query.Set("page", "Redemption_Code")
	query.Set("prop", "text|revid")
	query.Set("format", "json")
	query.Set("formatversion", "2")
	parsedURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request Fandom API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("Fandom API returned %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Fandom API response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("Fandom API response exceeds %d bytes", maxResponseSize)
	}
	var payload apiResponse
	if err := json.Unmarshal(body, &payload); err != nil {
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
	return parseHTML(payload.Parse.Text, s.URL(), now, payload.Parse.RevisionID)
}
