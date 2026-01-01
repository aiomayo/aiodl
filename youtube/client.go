package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	baseURL   = "https://www.youtube.com"
	userAgent = "com.google.android.youtube/19.09.37 (Linux; U; Android 11) gzip"
)

type Client struct {
	httpClient *http.Client
}

type Option func(*Client)

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

var (
	videoIDPattern    = regexp.MustCompile(`(?:v=|youtu\.be/|embed/|shorts/)([a-zA-Z0-9_-]{11})`)
	playlistIDPattern = regexp.MustCompile(`list=([a-zA-Z0-9_-]+)`)
)

func ExtractVideoID(url string) (string, error) {
	if len(url) == 11 {
		return url, nil
	}
	matches := videoIDPattern.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", ErrInvalidURL
	}
	return matches[1], nil
}

func ExtractPlaylistID(url string) (string, error) {
	matches := playlistIDPattern.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", ErrInvalidURL
	}
	return matches[1], nil
}

func IsYouTubeURL(url string) bool {
	return videoIDPattern.MatchString(url) || playlistIDPattern.MatchString(url)
}

func (c *Client) httpGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	return c.httpClient.Do(req)
}

func (c *Client) httpPost(ctx context.Context, url string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
