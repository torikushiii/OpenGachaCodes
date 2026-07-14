package hoyolab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"opengachacodes/internal/domain"
)

const (
	DefaultAPIURL   = "https://bbs-api-os.hoyolab.com/community/painter/wapi/circle/channel/guide/material?game_id=8"
	DefaultReferer  = "https://www.hoyolab.com/"
	maxResponseSize = 8 << 20
)

type Source struct {
	Client    *http.Client
	APIURL    string
	UserAgent string
	Now       func() time.Time
}

type apiResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		Modules []module `json:"modules"`
	} `json:"data"`
}

type module struct {
	ExchangeGroup *exchangeGroup `json:"exchange_group"`
}

type exchangeGroup struct {
	Bonuses []bonus `json:"bonuses"`
}

type bonus struct {
	ExchangeCode string      `json:"exchange_code"`
	CodeStatus   string      `json:"code_status"`
	Region       string      `json:"region"`
	Server       string      `json:"server"`
	IconBonuses  []iconBonus `json:"icon_bonuses"`
}

type iconBonus struct {
	BonusNum uint64 `json:"bonus_num"`
	IconURL  string `json:"icon_url"`
}

func (s *Source) ID() string       { return "zenless-hoyolab" }
func (s *Source) GameSlug() string { return "zenless" }
func (s *Source) URL() string {
	if strings.TrimSpace(s.APIURL) != "" {
		return s.APIURL
	}
	return DefaultAPIURL
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
		return nil, fmt.Errorf("create HoYoLAB request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-rpc-app_version", "4.8.0")
	request.Header.Set("x-rpc-client_type", "4")
	request.Header.Set("x-rpc-language", "en-us")
	request.Header.Set("Referer", DefaultReferer)

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request HoYoLAB API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("HoYoLAB API returned %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read HoYoLAB API response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("HoYoLAB API response exceeds %d bytes", maxResponseSize)
	}
	var payload apiResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode HoYoLAB API response: %w", err)
	}
	if payload.Retcode != 0 {
		return nil, fmt.Errorf("HoYoLAB API error %d: %s", payload.Retcode, payload.Message)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return parse(payload, s.URL(), now)
}
