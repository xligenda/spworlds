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

// ErrInvalidSignature is returned when a webhook body does not match its X-Body-Hash signature.
var ErrInvalidSignature = errors.New("invalid signature")

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
		return false, fmt.Errorf("failed to read request body: %w", err)
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
		return false, fmt.Errorf("invalid base64 signature: %w", err)
	}

	return hmac.Equal(c.computeSignature(bodyBytes), providedHash), nil
}

// VerifySignature reports whether signature (the X-Body-Hash header value) is a
// valid signature of body.
func (c *Client) VerifySignature(body []byte, signature string) bool {
	providedHash, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(c.computeSignature(body), providedHash)
}

func (c *Client) computeSignature(body []byte) []byte {
	mac := hmac.New(sha256.New, []byte(c.token))
	mac.Write(body)
	return mac.Sum(nil)
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

	out, err := ParsePaymentDataBytes(body)
	if err != nil {
		return nil, fmt.Errorf("ParsePaymentData: %w", err)
	}
	return out, nil
}

// ParsePaymentDataBytes parses a payment webhook body.
// Verify the body with VerifySignature before calling this.
func ParsePaymentDataBytes(body []byte) (*PaymentData, error) {
	var out PaymentData
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decoding body: %w", err)
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

	out, err := ParseReceivementDataBytes(body)
	if err != nil {
		return nil, fmt.Errorf("ParseReceivementData: %w", err)
	}
	return out, nil
}

// ParseReceivementDataBytes parses a card transaction webhook body.
// Verify the body with VerifySignature before calling this.
func ParseReceivementDataBytes(body []byte) (*ReceivementData, error) {
	var out ReceivementData
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decoding body: %w", err)
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
		return nil, fmt.Errorf("ParsePaymentDataValidated: %w", ErrInvalidSignature)
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
		return nil, fmt.Errorf("ParseReceivementDataValidated: %w", ErrInvalidSignature)
	}
	return c.ParseReceivementData(req)
}
