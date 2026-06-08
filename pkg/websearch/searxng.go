package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SearXNG 实现
type SearXNG struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

func NewSearXNG(endpoint, apiKey string) *SearXNG {
	return &SearXNG{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

type searxngResponse struct {
	Results []struct {
		Title   string  `json:"title"`
		URL     string  `json:"url"`
		Content string  `json:"content"`
		Score   float64 `json:"score"`
	} `json:"results"`
}

func (s *SearXNG) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	u, err := url.Parse(s.endpoint + "/search")
	if err != nil {
		return nil, fmt.Errorf("websearch: invalid endpoint: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	if limit > 0 {
		q.Set("number_of_results", fmt.Sprintf("%d", limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("websearch: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("websearch: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("websearch: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("websearch: API returned %d: %s", resp.StatusCode, string(body))
	}

	var sr searxngResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("websearch: parse response: %w", err)
	}

	results := make([]Result, 0, len(sr.Results))
	for _, r := range sr.Results {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
		})
	}
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
