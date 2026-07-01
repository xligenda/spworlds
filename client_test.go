package spworlds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func buildResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

func newTestClient(rt http.RoundTripper) *Client {
	return &Client{
		token:      "secret",
		apiKey:     "test-key",
		apiURL:     "https://api.test",
		httpClient: &http.Client{Transport: rt},
		userAgent:  "test-agent",
	}
}

func mustNewRequest(t *testing.T, method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	return req
}

func TestClient_Do(t *testing.T) {
	t.Run("Success Decodes JSON", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
			return buildResponse(http.StatusOK, []byte(`{"value":"some-value"}`)), nil
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		var dst struct {
			Value string `json:"value"`
		}

		err := c.do(req, &dst)
		require.NoError(t, err)
		assert.Equal(t, "some-value", dst.Value)
	})

	t.Run("Nil Destination Returns Nil", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, []byte(`{"value":"some-value"}`)), nil
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		err := c.do(req, nil)
		require.NoError(t, err)
	})

	t.Run("Non-2xx Parses API Error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusBadRequest, []byte(`{"message":"bad","error":"INVALID"}`)), nil
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		err := c.do(req, nil)
		require.Error(t, err)

		restErr, ok := err.(*RESTError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, restErr.StatusCode)
		assert.Equal(t, "INVALID", restErr.ErrorCode)
		assert.Equal(t, "bad", restErr.Message)
	})

	t.Run("Invalid JSON Returns Error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, []byte("not json")), nil
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		var dst struct {
			Value string `json:"value"`
		}

		err := c.do(req, &dst)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding response")
	})

	t.Run("Transport Error", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("transport failed")
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		err := c.do(req, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("Empty Body With Non-Nil Destination Returns Nil", func(t *testing.T) {
		c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return buildResponse(http.StatusOK, []byte("")), nil
		}))

		req := mustNewRequest(t, http.MethodGet, "https://api.test/endpoint")
		var dst struct {
			Value string `json:"value"`
		}

		err := c.do(req, &dst)
		require.NoError(t, err)
	})
}

func TestClient_NewRequest(t *testing.T) {
	t.Run("Payload Encoding Error", func(t *testing.T) {
		c := NewClient("id", "token", nil)
		type badPayload struct {
			Ch chan int `json:"ch"`
		}

		_, err := c.newRequest(context.Background(), http.MethodPost, "payments", badPayload{Ch: make(chan int)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "encoding request payload")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		c := NewClient("id", "token", nil)
		c.apiURL = ":"

		_, err := c.newRequest(context.Background(), http.MethodGet, "payments", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating request")
	})

	t.Run("Headers and Formatting", func(t *testing.T) {
		c := NewClient("id", "token", nil)
		c.apiURL = "https://api.test"

		options := CreatePaymentOptions{RedirectURL: "https://discord.com"}
		req, err := c.newRequest(context.Background(), http.MethodPost, "payments", options)
		require.NoError(t, err)

		assert.Equal(t, "https://api.test/payments", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer "+c.apiKey, req.Header.Get("Authorization"))

		var payload CreatePaymentOptions
		err = json.NewDecoder(req.Body).Decode(&payload)
		require.NoError(t, err)
		assert.Equal(t, "https://discord.com", payload.RedirectURL)
	})
}

func TestParseAPIError(t *testing.T) {
	t.Run("Valid JSON Response", func(t *testing.T) {
		err := parseAPIError(http.StatusPaymentRequired, []byte(`{"message":"need funds","error":"NOT_ENOUGH"}`))
		restErr, ok := err.(*RESTError)
		require.True(t, ok)
		assert.Equal(t, http.StatusPaymentRequired, restErr.StatusCode)
		assert.Equal(t, "NOT_ENOUGH", restErr.ErrorCode)
		assert.Equal(t, "need funds", restErr.Message)
	})

	t.Run("Fallback to Plain Text", func(t *testing.T) {
		err := parseAPIError(http.StatusInternalServerError, []byte("server down"))
		restErr, ok := err.(*RESTError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, restErr.StatusCode)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError), restErr.ErrorCode)
	})
}

func TestClient_HTTPMethods(t *testing.T) {
	t.Run("Get Propagates NewRequest Error", func(t *testing.T) {
		c := newTestClient(nil)
		c.apiURL = ":"

		err := c.get(context.Background(), "payments", &struct{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating request")
	})

	t.Run("Post Propagates NewRequest Error", func(t *testing.T) {
		c := newTestClient(nil)
		c.apiURL = ":"

		err := c.post(context.Background(), "payments", CreatePaymentOptions{}, &struct{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating request")
	})

	t.Run("Put Propagates NewRequest Error", func(t *testing.T) {
		c := newTestClient(nil)
		c.apiURL = ":"

		err := c.put(context.Background(), "webhooks", UpdateWebhookOptions{}, &struct{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating request")
	})
}
