package spworlds

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const DEFAULT_APIURL = "https://spworlds.ru/api/public"

type RESTError struct {
	Message    string `json:"message"`
	ErrorCode  string `json:"error"`
	StatusCode int    `json:"statusCode"`
}

func (e *RESTError) Error() string {
	return fmt.Sprintf("API Error %d: %s - %s", e.StatusCode, e.ErrorCode, e.Message)
}

type Client struct {
	sync.Mutex

	// Для использования API вам надо знать ID и токен для карты, с которой вы хотите совершить действие.
	// Получить ID и токен можно на странице "Кошелёк", в секции "Поделиться картой".
	// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
	ID string

	// Для использования API вам надо знать ID и токен для карты, с которой вы хотите совершить действие.
	// Получить ID и токен можно на странице "Кошелёк", в секции "Поделиться картой".
	// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
	Token string

	// base64 закодированная строка "ID:TOKEN", где ID - ID вашей карты, TOKEN - токен от нее.
	// https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F
	APIKey string

	// The user agent used
	UserAgent string

	// URL публичного API spworlds
	// Стандартное значение - https://spworlds.ru/api/public
	APIURL string

	// The http client used
	HttpClient *http.Client
}

func NewClient(id string, token string) *Client {
	return &Client{
		ID:         id,
		Token:      token,
		APIKey:     base64.StdEncoding.EncodeToString([]byte(id + ":" + token)),
		HttpClient: &http.Client{},
		APIURL:     DEFAULT_APIURL,
		UserAgent:  "spworlds-go-client/1.0",
	}
}

func (c *Client) doRequest(req *http.Request, response interface{}) error {
	c.Lock()
	defer c.Unlock()

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseError(resp.StatusCode, respBody)
	}

	return c.parseResponse(respBody, response)
}

func (c *Client) parseError(statusCode int, respBody []byte) error {
	var restError RESTError
	if err := json.Unmarshal(respBody, &restError); err == nil {
		restError.StatusCode = statusCode
		return &restError
	}

	return &RESTError{
		StatusCode: statusCode,
		ErrorCode:  fmt.Sprintf("HTTP %d", statusCode),
		Message:    string(respBody),
	}
}

func (c *Client) parseResponse(respBody []byte, response any) error {
	if response == nil {
		return nil
	}

	if len(respBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) parseRequest(method, endpoint string, payload interface{}) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s/%s", c.APIURL, endpoint)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	return req, nil
}
