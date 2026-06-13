// Package comick is the library behind the comick command line:
// the HTTP client, request shaping, and the typed data models for comick.io.
//
// The API at api.comick.io is open and requires no authentication for reads.
// The client paces requests and retries 429 and 5xx responses.
package comick

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrNotFound is returned when the API returns an empty or null result.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.comick.io",
		UserAgent: "comick-cli/0.1",
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   15 * time.Second,
	}
}

// Client talks to the comick.io API over HTTP.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	rate       time.Duration
	retries    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, false, fmt.Errorf("http 404")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "null" || trimmed == "" {
		return ErrNotFound
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// ─── API methods ──────────────────────────────────────────────────────────────

// Search returns comics matching q. sort may be "view", "new", "trending",
// "follow", or "rating". country and typ are optional filter strings.
func (c *Client) Search(ctx context.Context, q string, limit int, sort, country, typ string) ([]Comic, error) {
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("q", q)
	params.Set("limit", strconv.Itoa(limit))
	if sort != "" {
		params.Set("sort", sort)
	}
	if country != "" {
		params.Set("country", country)
	}
	if typ != "" {
		params.Set("type", typ)
	}
	rawURL := c.baseURL + "/v1.0/search/?" + params.Encode()
	var wcs []wireComic
	if err := c.getJSON(ctx, rawURL, &wcs); err != nil {
		return nil, err
	}
	if len(wcs) == 0 {
		return nil, ErrNotFound
	}
	out := make([]Comic, len(wcs))
	for i, w := range wcs {
		out[i] = wireComicToComic(w)
	}
	return out, nil
}

// Trending returns currently trending comics.
func (c *Client) Trending(ctx context.Context, limit int) ([]Comic, error) {
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("sort", "trending")
	params.Set("limit", strconv.Itoa(limit))
	rawURL := c.baseURL + "/v1.0/search/?" + params.Encode()
	var wcs []wireComic
	if err := c.getJSON(ctx, rawURL, &wcs); err != nil {
		return nil, err
	}
	if len(wcs) == 0 {
		return nil, ErrNotFound
	}
	out := make([]Comic, len(wcs))
	for i, w := range wcs {
		out[i] = wireComicToComic(w)
	}
	return out, nil
}

// New returns the most recently added comics.
func (c *Client) New(ctx context.Context, limit int) ([]Comic, error) {
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("sort", "new")
	params.Set("limit", strconv.Itoa(limit))
	rawURL := c.baseURL + "/v1.0/search/?" + params.Encode()
	var wcs []wireComic
	if err := c.getJSON(ctx, rawURL, &wcs); err != nil {
		return nil, err
	}
	if len(wcs) == 0 {
		return nil, ErrNotFound
	}
	out := make([]Comic, len(wcs))
	for i, w := range wcs {
		out[i] = wireComicToComic(w)
	}
	return out, nil
}

// GetComic returns full metadata for one comic by hid or slug.
func (c *Client) GetComic(ctx context.Context, hid string) (Comic, error) {
	rawURL := c.baseURL + "/comic/" + url.PathEscape(hid) + "/"
	var d wireComicDetail
	if err := c.getJSON(ctx, rawURL, &d); err != nil {
		if strings.Contains(err.Error(), "http 404") {
			return Comic{}, ErrNotFound
		}
		return Comic{}, err
	}
	if d.Comic.HID == "" && d.Comic.Slug == "" {
		return Comic{}, ErrNotFound
	}
	return wireComicDetailToComic(d), nil
}

// GetChapters returns chapters for a comic. lang="" skips the lang filter.
// asc=true returns chapters in ascending order (oldest first).
func (c *Client) GetChapters(ctx context.Context, hid string, limit, page int, lang string, asc bool) ([]Chapter, error) {
	if limit <= 0 {
		limit = 60
	}
	if page <= 0 {
		page = 1
	}
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("limit", strconv.Itoa(limit))
	if lang != "" {
		params.Set("lang", lang)
	}
	if asc {
		params.Set("chap-order", "0")
	} else {
		params.Set("chap-order", "1")
	}
	rawURL := c.baseURL + "/comic/" + url.PathEscape(hid) + "/chapters?" + params.Encode()
	var resp wireChaptersResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}
	if len(resp.Chapters) == 0 {
		return nil, ErrNotFound
	}
	out := make([]Chapter, len(resp.Chapters))
	for i, w := range resp.Chapters {
		out[i] = wireChapterToChapter(w)
	}
	return out, nil
}
