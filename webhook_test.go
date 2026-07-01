package spworlds

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorReadCloser struct{ err error }

func (e errorReadCloser) Read(p []byte) (int, error) { return 0, e.err }
func (e errorReadCloser) Close() error               { return nil }

func makeSignedRequest(t *testing.T, token string, body any) *http.Request {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "https://example.com/webhook", bytes.NewReader(data))
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte(token))
	mac.Write(data)
	req.Header.Set("X-Body-Hash", base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	return req
}

func TestValidateRequest_Success(t *testing.T) {
	c := &Client{token: "secret"}
	req := makeSignedRequest(t, "secret", map[string]any{"payer": "5opka"})

	ok, err := c.ValidateRequest(req)
	require.NoError(t, err)
	assert.True(t, ok)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"payer":"5opka"}`, string(body))
}

func TestValidateRequest_MissingHeader(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(`{"payer":"5opka"}`)))
	require.NoError(t, err)

	ok, err := c.ValidateRequest(req)
	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing signature header")
}

func TestValidateRequest_InvalidBase64(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(`{"payer":"5opka"}`)))
	require.NoError(t, err)
	req.Header.Set("X-Body-Hash", "!!!")

	ok, err := c.ValidateRequest(req)
	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base64 signature")
}

func TestValidateRequest_WrongSignature(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(`{"payer":"5opka"}`)))
	require.NoError(t, err)

	wrongHash := base64.StdEncoding.EncodeToString([]byte("wrong"))
	req.Header.Set("X-Body-Hash", wrongHash)

	ok, err := c.ValidateRequest(req)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateRequest_ReadError(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = errorReadCloser{err: errors.New("read fail")}

	ok, err := c.ValidateRequest(req)
	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read request body")
}

func TestParsePaymentData_ValidAndInvalid(t *testing.T) {
	c := &Client{token: "secret"}

	requestBody := `{"payer":"5opka","amount":123,"data":"payload"}`
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(requestBody)))
	require.NoError(t, err)

	payment, err := c.ParsePaymentData(req)
	require.NoError(t, err)
	assert.Equal(t, "5opka", payment.Payer)
	assert.Equal(t, 123, payment.Amount)
	assert.Equal(t, "payload", payment.Payload)

	badReq, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte("not json")))
	require.NoError(t, err)

	_, err = c.ParsePaymentData(badReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParsePaymentData: decoding body")
}

func TestParsePaymentData_ReadError(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = errorReadCloser{err: errors.New("read fail")}

	_, err = c.ParsePaymentData(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParsePaymentData: failed to read body")
}

func TestParseReceivementData_ValidAndInvalid(t *testing.T) {
	c := &Client{token: "secret"}

	requestBody := `{
		"id":"tx-123",
		"amount":400,
		"type":"payment",
		"sender":{"username":"oster","number":"1234"},
		"receiver":{"username":"5opka","number":"5678"},
		"comment":"thanks",
		"createdAt":"2026-07-01T12:00:00.000Z"
	}`
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(requestBody)))
	require.NoError(t, err)

	received, err := c.ParseReceivementData(req)
	require.NoError(t, err)
	assert.Equal(t, "tx-123", received.ID)
	assert.Equal(t, 400, received.Amount)
	assert.Equal(t, "payment", received.Type)
	require.NotNil(t, received.Sender)
	assert.Equal(t, "oster", *received.Sender.Username)
	require.NotNil(t, received.Receiver)
	assert.Equal(t, "5opka", *received.Receiver.Username)
	assert.Equal(t, "thanks", received.Comment)

	badReq, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte("not json")))
	require.NoError(t, err)

	_, err = c.ParseReceivementData(badReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParseReceivementData: decoding body")
}

func TestParseReceivementData_ReadError(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = errorReadCloser{err: errors.New("read fail")}

	_, err = c.ParseReceivementData(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParseReceivementData: failed to read body")
}

func TestParsePaymentDataValidated(t *testing.T) {
	c := &Client{token: "secret"}
	req := makeSignedRequest(t, "secret", map[string]any{"payer": "5opka", "amount": 123, "data": "payload"})

	payment, err := c.ParsePaymentDataValidated(req)
	require.NoError(t, err)
	assert.Equal(t, "5opka", payment.Payer)
}

func TestParsePaymentDataValidated_InvalidSignature(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(`{"payer":"5opka"}`)))
	require.NoError(t, err)
	req.Header.Set("X-Body-Hash", base64.StdEncoding.EncodeToString([]byte("wrong")))

	_, err = c.ParsePaymentDataValidated(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestParsePaymentDataValidated_ReadError(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = errorReadCloser{err: errors.New("read fail")}
	req.Header = make(http.Header)
	req.Header.Set("X-Body-Hash", base64.StdEncoding.EncodeToString([]byte("wrong")))

	_, err = c.ParsePaymentDataValidated(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParsePaymentDataValidated: validating request")
}

func TestParseReceivementDataValidated(t *testing.T) {
	c := &Client{token: "secret"}
	req := makeSignedRequest(t, "secret", map[string]any{"id": "tx-1", "amount": 10, "type": "payment"})

	received, err := c.ParseReceivementDataValidated(req)
	require.NoError(t, err)
	assert.Equal(t, "tx-1", received.ID)
}

func TestParseReceivementDataValidated_InvalidSignature(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(`{"id":"tx-1"}`)))
	require.NoError(t, err)
	req.Header.Set("X-Body-Hash", base64.StdEncoding.EncodeToString([]byte("wrong")))

	_, err = c.ParseReceivementDataValidated(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestParseReceivementDataValidated_ReadError(t *testing.T) {
	c := &Client{token: "secret"}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = errorReadCloser{err: errors.New("read fail")}
	req.Header = make(http.Header)
	req.Header.Set("X-Body-Hash", base64.StdEncoding.EncodeToString([]byte("wrong")))

	_, err = c.ParseReceivementDataValidated(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ParseReceivementDataValidated: validating request")
}
