package avatars

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xligenda/spworlds"
	"github.com/xligenda/spworlds/ratelimit"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func buildResponse(status int, contentType string, body []byte) *http.Response {
	header := make(http.Header)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     header,
	}
}

func newTestClient(rt http.RoundTripper) *Client {
	return &Client{
		baseURL:    "https://avatars.test",
		httpClient: &http.Client{Transport: rt},
		userAgent:  "test-agent",
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient()

	assert.Equal(t, DefaultBaseURL, c.baseURL)
	assert.Equal(t, spworlds.DefaultUserAgent, c.userAgent)
	require.NotNil(t, c.httpClient)
	assert.Equal(t, 15*time.Second, c.httpClient.Timeout)
	require.NotNil(t, c.limiter)
}

func TestNewClient_ConfigOverride(t *testing.T) {
	customClient := &http.Client{Timeout: 1 * time.Second}
	baseURL := "https://example.com/avatars"
	userAgent := "test-agent/0.1"
	limiter := ratelimit.NewRateLimiter(10)

	c := NewClient(
		WithBaseURL(baseURL),
		WithHTTPClient(customClient),
		WithUserAgent(userAgent),
		WithRateLimiter(limiter),
	)

	assert.Equal(t, baseURL, c.baseURL)
	assert.Equal(t, customClient, c.httpClient)
	assert.Equal(t, userAgent, c.userAgent)
	assert.Equal(t, limiter, c.limiter)
}

func TestNewClient_OptionsIgnoreZeroValues(t *testing.T) {
	c := NewClient(
		WithBaseURL(""),
		WithHTTPClient(nil),
		WithUserAgent(""),
		WithRateLimiter(nil),
	)

	assert.Equal(t, DefaultBaseURL, c.baseURL)
	assert.Equal(t, spworlds.DefaultUserAgent, c.userAgent)
	require.NotNil(t, c.httpClient)
	require.NotNil(t, c.limiter)
}

func TestNewClient_NilOptionIsSkipped(t *testing.T) {
	c := NewClient(nil, WithBaseURL("https://example.com"))
	assert.Equal(t, "https://example.com", c.baseURL)
}

func TestClient_Do(t *testing.T) {
	t.Run("success returns response", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://avatars.test/head/5opka?width=128", req.URL.String())
			assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
			return buildResponse(http.StatusOK, "image/png", []byte("png-bytes")), nil
		}))

		resp, err := c.do(context.Background(), "5opka", Head, 128)
		require.NoError(t, err)
		require.NotNil(t, resp)
		// response body close error is not actionable
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("skips User-Agent header when empty", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Empty(t, req.Header.Get("User-Agent"))
			return buildResponse(http.StatusOK, "image/png", []byte("data")), nil
		}))
		c.userAgent = ""

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		require.NoError(t, err)
		// response body close error is not actionable
		defer func() { _ = resp.Body.Close() }()
	})

	t.Run("request creation error", func(t *testing.T) {
		c := newTestClient(nil)
		c.baseURL = "://bad-url"

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		if resp != nil {
			// response body close error is not actionable
			defer func() { _ = resp.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating request")
	})

	t.Run("transport error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}))

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		if resp != nil {
			// response body close error is not actionable
			defer func() { _ = resp.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("non-2xx status with body wraps message", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusNotFound, "text/plain", []byte("player not found")), nil
		}))

		resp, err := c.do(context.Background(), "unknown", Head, 0)
		if resp != nil {
			// response body close error is not actionable
			defer func() { _ = resp.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Contains(t, err.Error(), `"unknown"`)
		assert.Contains(t, err.Error(), "player not found")
	})

	t.Run("non-2xx status without body", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusInternalServerError, "", nil), nil
		}))

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		if resp != nil {
			// response body close error is not actionable
			defer func() { _ = resp.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
		assert.NotContains(t, err.Error(), ":")
	})

	t.Run("unexpected content type", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, "application/json", []byte(`{"error":"nope"}`)), nil
		}))

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		if resp != nil {
			// response body close error is not actionable
			defer func() { _ = resp.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected content-type")
		assert.Contains(t, err.Error(), `"5opka"`)
	})

	t.Run("uses rate limiter and propagates cancellation", func(t *testing.T) {
		limiter := ratelimit.NewRateLimiterWithDelta(1, time.Hour)
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, "image/png", []byte("data")), nil
		}))
		c.limiter = limiter

		resp, err := c.do(context.Background(), "5opka", Head, 0)
		require.NoError(t, err)
		// response body close error is not actionable
		_ = resp.Body.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		resp2, err := c.do(ctx, "5opka", Head, 0)
		if resp2 != nil {
			// response body close error is not actionable
			defer func() { _ = resp2.Body.Close() }()
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limiter")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestClient_Fetch(t *testing.T) {
	t.Run("returns body bytes", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, "image/png", []byte("png-bytes")), nil
		}))

		data, err := c.Fetch(context.Background(), "5opka", Head, 128)
		require.NoError(t, err)
		assert.Equal(t, []byte("png-bytes"), data)
	})

	t.Run("propagates do error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusBadRequest, "text/plain", []byte("bad player")), nil
		}))

		_, err := c.Fetch(context.Background(), "5opka", Head, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad player")
	})

	t.Run("read error surfaced", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := buildResponse(http.StatusOK, "image/png", nil)
			resp.Body = io.NopCloser(errorReader{})
			return resp, nil
		}))

		_, err := c.Fetch(context.Background(), "5opka", Head, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading response body")
	})
}

func TestClient_FetchTo(t *testing.T) {
	t.Run("copies body to writer", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, "image/png", []byte("png-bytes")), nil
		}))

		var buf bytes.Buffer
		n, err := c.FetchTo(context.Background(), &buf, "5opka", Head, 128)
		require.NoError(t, err)
		assert.Equal(t, int64(len("png-bytes")), n)
		assert.Equal(t, "png-bytes", buf.String())
	})

	t.Run("propagates do error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusForbidden, "text/plain", []byte("forbidden")), nil
		}))

		var buf bytes.Buffer
		_, err := c.FetchTo(context.Background(), &buf, "5opka", Head, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "forbidden")
	})

	t.Run("copy error surfaced", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := buildResponse(http.StatusOK, "image/png", nil)
			resp.Body = io.NopCloser(errorReader{})
			return resp, nil
		}))

		var buf bytes.Buffer
		_, err := c.FetchTo(context.Background(), &buf, "5opka", Head, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "copying response body")
	})
}

// errorReader always fails to read, used to exercise io error paths.
type errorReader struct{}

func (errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestClient_PartMethods(t *testing.T) {
	tests := []struct {
		name string
		call func(c *Client, ctx context.Context, player string, width int) ([]byte, error)
		part Part
	}{
		{"Head", (*Client).Head, Head},
		{"Front", (*Client).Front, Front},
		{"Body", (*Client).Body, Body},
		{"Bust", (*Client).Bust, Bust},
		{"Full", (*Client).Full, Full},
		{"FrontFull", (*Client).FrontFull, FrontFull},
		{"Face", (*Client).Face, Face},
		{"Skin", (*Client).Skin, Skin},
		{"Cape", (*Client).Cape, Cape},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				gotPath = req.URL.Path
				return buildResponse(http.StatusOK, "image/png", []byte("data")), nil
			}))

			data, err := tt.call(c, context.Background(), "5opka", 64)
			require.NoError(t, err)
			assert.Equal(t, []byte("data"), data)
			assert.Equal(t, "/"+string(tt.part)+"/5opka", gotPath)
		})
	}
}

func TestClient_Integration_WithHTTPTestServer(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path == "/cape/nocape" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-png-data"))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithBaseURL(srv.URL))

	t.Run("head fetch succeeds", func(t *testing.T) {
		data, err := c.Head(context.Background(), "5opka", 128)
		require.NoError(t, err)
		assert.Equal(t, "fake-png-data", string(data))
	})

	t.Run("missing cape returns error", func(t *testing.T) {
		_, err := c.Cape(context.Background(), "nocape", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	assert.GreaterOrEqual(t, atomic.LoadInt32(&hits), int32(2))
}
