package avatars

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xligenda/spworlds"
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

type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:   DefaultBaseURL,
		userAgent: spworlds.DefaultUserAgent,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) doRequest(ctx context.Context, uuid string, part Part, width int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(uuid, part, width), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
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
			return nil, fmt.Errorf("avatar service returned status %d for %q: %s", resp.StatusCode, uuid, bytes.TrimSpace(body))
		}
		return nil, fmt.Errorf("avatar service returned status %d for %q", resp.StatusCode, uuid)
	}

	return resp, nil
}

func (c *Client) Fetch(ctx context.Context, uuid string, part Part, width int) ([]byte, error) {
	resp, err := c.doRequest(ctx, uuid, part, width)
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

func (c *Client) FetchTo(ctx context.Context, dst io.Writer, uuid string, part Part, width int) (int64, error) {
	resp, err := c.doRequest(ctx, uuid, part, width)
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
