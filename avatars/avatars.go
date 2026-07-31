package avatars

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xligenda/spworlds"
	"github.com/xligenda/spworlds/ratelimit"
)

const (
	DefaultBaseURL = "https://avatars.spworlds.ru"
)

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		if userAgent != "" {
			c.userAgent = userAgent
		}
	}
}

func WithRateLimiter(limiter *ratelimit.RateLimiter) Option {
	return func(c *Client) {
		if limiter != nil {
			c.limiter = limiter
		}
	}
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string

	limiter *ratelimit.RateLimiter
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:   DefaultBaseURL,
		userAgent: spworlds.DefaultUserAgent,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		limiter: ratelimit.NewRateLimiter(100),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(c)
	}

	return c
}

func (c *Client) do(ctx context.Context, player string, part Part, width int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(player, part, width), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	if c.limiter != nil {
		if err := c.limiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// response body close error is not actionable
		defer resp.Body.Close() //nolint:errcheck
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if len(body) > 0 {
			return nil, fmt.Errorf("avatar service returned status %d for %q: %s", resp.StatusCode, player, bytes.TrimSpace(body))
		}
		return nil, fmt.Errorf("avatar service returned status %d for %q", resp.StatusCode, player)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		// response body close error is not actionable
		defer resp.Body.Close() //nolint:errcheck
		return nil, fmt.Errorf("unexpected content-type %q for player %q", ct, player)
	}

	return resp, nil
}

func (c *Client) Fetch(ctx context.Context, player string, part Part, width int) ([]byte, error) {
	resp, err := c.do(ctx, player, part, width)
	if err != nil {
		return nil, err
	}
	// response body close error is not actionable
	defer resp.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return data, nil
}

func (c *Client) FetchTo(ctx context.Context, dst io.Writer, player string, part Part, width int) (int64, error) {
	resp, err := c.do(ctx, player, part, width)
	if err != nil {
		return 0, err
	}
	// response body close error is not actionable
	defer resp.Body.Close() //nolint:errcheck

	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return n, fmt.Errorf("copying response body: %w", err)
	}
	return n, nil
}

// head, 2D render
func (c *Client) Head(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Head, width)
}

// front bust, 2D render
func (c *Client) Front(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Front, width)
}

// front bust, 2D render
func (c *Client) Body(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Body, width)
}

// bust, 3D render
func (c *Client) Bust(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Bust, width)
}

// full body, 3D render
func (c *Client) Full(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Full, width)
}

// full body, 2D render
func (c *Client) FrontFull(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, FrontFull, width)
}

// face without skin second layer, 2D render
func (c *Client) Face(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Face, width)
}

// raw skin texture, PNG
func (c *Client) Skin(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Skin, width)
}

// cape texture, PNG
// if player is not wearing any cape, return error code 404
func (c *Client) Cape(ctx context.Context, player string, width int) ([]byte, error) {
	return c.Fetch(ctx, player, Cape, width)
}
