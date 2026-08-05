package spworlds

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxWebhookBodyBytes bounds how much of a webhook request body is read into
// memory. It is enforced before the signature is verified, so it must stay
// generous enough for legitimate payloads while capping worst-case memory use.
const maxWebhookBodyBytes = 64 * 1024

// readLimitedBody reads up to maxWebhookBodyBytes+1 bytes from r and errors
// out if the body turns out to be larger than the limit.
func readLimitedBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxWebhookBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxWebhookBodyBytes {
		return nil, fmt.Errorf("body exceeds %d bytes limit", maxWebhookBodyBytes)
	}
	return body, nil
}

// ValidateRequest returns true if the request signature is valid.
func (c *Client) ValidateRequest(req *http.Request) (bool, error) {
	bodyBytes, err := readLimitedBody(req.Body)
	if err != nil {
		return false, errors.New("failed to read request body: " + err.Error())
	}
	// request body close error is not actionable
	_ = req.Body.Close()

	// restore the body so downstream handlers can read it
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	hashHeader := req.Header.Get("X-Body-Hash")
	if hashHeader == "" {
		return false, errors.New("missing signature header")
	}

	providedHash, err := base64.StdEncoding.DecodeString(hashHeader)
	if err != nil {
		return false, errors.New("invalid base64 signature: " + err.Error())
	}

	mac := hmac.New(sha256.New, []byte(c.token))
	mac.Write(bodyBytes)
	computedHash := mac.Sum(nil)

	if !hmac.Equal(computedHash, providedHash) {
		return false, nil
	}
	return true, nil
}

type PaymentData struct {
	// Ник игрока, который совершил оплату.
	Payer string `json:"payer"`
	// Стоимость покупки.
	Amount int `json:"amount"`
	// Данные, которые вы отдали при создании запроса на оплату.
	Payload string `json:"data"`
}

// Validate the request with ValidateRequest before calling this.
func (c *Client) ParsePaymentData(req *http.Request) (*PaymentData, error) {
	body, err := readLimitedBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("ParsePaymentData: failed to read body: %w", err)
	}
	// request body close error is not actionable
	defer req.Body.Close() //nolint:errcheck

	var out PaymentData
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("ParsePaymentData: decoding body: %w", err)
	}
	return &out, nil
}

type ReceivementData struct {
	// Уникальный ID транзакции.
	ID string `json:"id"`
	// Сумма транзакции.
	Amount int `json:"amount"`
	// Тип транзакции.
	Type   string `json:"type"`
	Sender *struct {
		// Ник отправителя (если есть).
		Username *string `json:"username"`
		// Номер карты отправителя (если есть).
		Number *string `json:"number"`
	} `json:"sender"`
	Receiver *struct {
		// Ник получателя (если есть).
		Username *string `json:"username"`
		// Номер карты получателя (если есть).
		Number *string `json:"number"`
	} `json:"receiver"`
	// Комментарий к транзакции.
	Comment string `json:"comment"`
	// Дата создания транзакции.
	CreatedAt string `json:"createdAt"`
}

// Validate the request with ValidateRequest before calling this.
func (c *Client) ParseReceivementData(req *http.Request) (*ReceivementData, error) {
	body, err := readLimitedBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("ParseReceivementData: failed to read body: %w", err)
	}
	// request body close error is not actionable
	defer req.Body.Close() //nolint:errcheck

	var out ReceivementData
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("ParseReceivementData: decoding body: %w", err)
	}
	return &out, nil
}

// ParsePaymentDataValidated validates the request signature and parses the payment webhook body.
// It is a convenience wrapper around ValidateRequest and ParsePaymentData.
func (c *Client) ParsePaymentDataValidated(req *http.Request) (*PaymentData, error) {
	ok, err := c.ValidateRequest(req)
	if err != nil {
		return nil, fmt.Errorf("ParsePaymentDataValidated: validating request: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("ParsePaymentDataValidated: invalid signature")
	}
	return c.ParsePaymentData(req)
}

// ParseReceivementDataValidated validates the request signature and parses the receivement webhook body.
// It is a convenience wrapper around ValidateRequest and ParseReceivementData.
func (c *Client) ParseReceivementDataValidated(req *http.Request) (*ReceivementData, error) {
	ok, err := c.ValidateRequest(req)
	if err != nil {
		return nil, fmt.Errorf("ParseReceivementDataValidated: validating request: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("ParseReceivementDataValidated: invalid signature")
	}
	return c.ParseReceivementData(req)
}
