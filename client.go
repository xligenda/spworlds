package spworlds

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultAPIURL = "https://spworlds.ru/api/public"
	VERSION       = "1.2"
)

// RESTError отображает ошибку SPWorlds API.
type RESTError struct {
	Message    string `json:"message"`
	ErrorCode  string `json:"error"`
	StatusCode int    `json:"statusCode"`
}

func (e *RESTError) Error() string {
	return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.ErrorCode, e.Message)
}

type ClientConfig struct {
	APIURL     string
	UserAgent  string
	HTTPClient *http.Client
}

type Client struct {
	// Для использования API вам надо знать ID и токен для карты, с которой вы хотите совершить действие.
	// Получить ID и токен можно на странице "Кошелёк", в секции "Поделиться картой".
	// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
	token string

	// base64 закодированная строка "ID:TOKEN", где ID - ID вашей карты, TOKEN - токен от нее.
	// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
	apiKey string

	// URL публичного API spworlds
	// Стандартное значение - https://spworlds.ru/api/public
	apiURL string

	httpClient *http.Client
	userAgent  string
}

// Для использования API вам надо знать ID и токен для карты, с которой вы хотите совершить действие.
// Получить ID и токен можно на странице "Кошелёк", в секции "Поделиться картой".
// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
func NewClient(id, token string, cfg *ClientConfig) *Client {
	c := &Client{
		token:     token,
		apiKey:    base64.StdEncoding.EncodeToString([]byte(id + ":" + token)),
		apiURL:    DefaultAPIURL,
		userAgent: "spworlds-go (github.com/xligenda/spworlds, v" + VERSION + ")",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	if cfg != nil {
		if cfg.APIURL != "" {
			c.apiURL = cfg.APIURL
		}
		if cfg.UserAgent != "" {
			c.userAgent = cfg.UserAgent
		}
		if cfg.HTTPClient != nil {
			c.httpClient = cfg.HTTPClient
		}
	}
	return c
}

func (c *Client) do(req *http.Request, dst any) error {
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, body)
	}

	if dst == nil || len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	url := c.apiURL + "/" + endpoint

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding request payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return req, nil
}

func (c *Client) get(ctx context.Context, endpoint string, dst any) error {
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	return c.do(req, dst)
}

func (c *Client) post(ctx context.Context, endpoint string, payload, dst any) error {
	req, err := c.newRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}
	return c.do(req, dst)
}

func (c *Client) put(ctx context.Context, endpoint string, payload, dst any) error {
	req, err := c.newRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}
	return c.do(req, dst)
}

func parseAPIError(statusCode int, body []byte) error {
	var e RESTError
	if json.Unmarshal(body, &e) == nil && e.ErrorCode != "" {
		e.StatusCode = statusCode
		return &e
	}
	return &RESTError{
		StatusCode: statusCode,
		ErrorCode:  http.StatusText(statusCode),
		Message:    string(body),
	}
}
