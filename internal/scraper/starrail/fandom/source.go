package fandom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"opengachacodes/internal/domain"
	"strings"
	"time"
)

const (
	DefaultAPIURL   = "https://honkai-star-rail.fandom.com/api.php"
	DefaultPageURL  = "https://honkai-star-rail.fandom.com/wiki/Redemption_Code"
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

func (s *Source) ID() string       { return "starrail-fandom" }
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
	api := s.APIURL
	if api == "" {
		api = DefaultAPIURL
	}
	ua := strings.TrimSpace(s.UserAgent)
	if ua == "" {
		return nil, fmt.Errorf("user agent is required")
	}
	u, e := url.Parse(api)
	if e != nil {
		return nil, fmt.Errorf("parse API URL: %w", e)
	}
	q := u.Query()
	q.Set("action", "parse")
	q.Set("page", "Redemption_Code")
	q.Set("prop", "text|revid")
	q.Set("format", "json")
	q.Set("formatversion", "2")
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return nil, fmt.Errorf("create request: %w", e)
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("request Fandom API: %w", e)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Fandom API returned %s", resp.Status)
	}
	body, e := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if e != nil {
		return nil, fmt.Errorf("read Fandom API response: %w", e)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("Fandom API response exceeds %d bytes", maxResponseSize)
	}
	var p apiResponse
	if e := json.Unmarshal(body, &p); e != nil {
		return nil, fmt.Errorf("decode Fandom API response: %w", e)
	}
	if p.Error != nil {
		return nil, fmt.Errorf("Fandom API error %s: %s", p.Error.Code, p.Error.Info)
	}
	if strings.TrimSpace(p.Parse.Text) == "" {
		return nil, fmt.Errorf("Fandom API response has no parsed HTML")
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return parseHTML(p.Parse.Text, s.URL(), now, p.Parse.RevisionID)
}
