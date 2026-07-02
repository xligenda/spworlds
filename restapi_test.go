package spworlds

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientCard_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "https://api.test/card", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`{"balance":100,"webhook":"https://webhook.com"}`)), nil
	}))

	card, err := c.ClientCard(context.Background())
	require.NoError(t, err)
	require.NotNil(t, card)
	assert.Equal(t, 100, card.Balance)
	require.NotNil(t, card.Webhook)
	assert.Equal(t, "https://webhook.com", *card.Webhook)
}

func TestClientCard_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.ClientCard(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClientCard:")
	assert.Contains(t, err.Error(), "API error 400")
}

func TestUser_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "https://api.test/users/888016163844534372", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`{"uuid":"b963413a-b97f-4124-aebf-9a1eefd0b144","username":"5opka"}`)), nil
	}))

	user, err := c.User(context.Background(), "888016163844534372")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "b963413a-b97f-4124-aebf-9a1eefd0b144", user.UUID)
	assert.Equal(t, "5opka", user.Username)
}

func TestUser_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusNotFound, []byte(`{"message":"not found","error":"NOT_FOUND"}`)), nil
	}))

	_, err := c.User(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "User(missing):")
	assert.Contains(t, err.Error(), "API error 404")
}

func TestUserCards_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "https://api.test/accounts/5opka/cards", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`[{"id":"c-1","name":"Card One","number":"1111"}]`)), nil
	}))

	cards, err := c.UserCards(context.Background(), "5opka")
	require.NoError(t, err)
	assert.Len(t, cards, 1)
	assert.Equal(t, "c-1", *cards[0].ID)
	assert.Equal(t, "Card One", cards[0].Name)
}

func TestUserCards_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.UserCards(context.Background(), "5opka")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UserCards(5opka):")
	assert.Contains(t, err.Error(), "API error 400")
}

func TestMe_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "https://api.test/accounts/me", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`{"id":"b963413a-b97f-4124-aebf-9a1eefd0b144","username":"5opka"}`)), nil
	}))

	user, err := c.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "5opka", user.Username)
}

func TestMe_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.Me(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Me:")
	assert.Contains(t, err.Error(), "API error 400")
}

func TestCreatePayment_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://api.test/payments", req.URL.String())
		var payload CreatePaymentOptions
		err := json.NewDecoder(req.Body).Decode(&payload)
		require.NoError(t, err)
		assert.Equal(t, "https://discord.com", payload.RedirectURL)
		return buildResponse(http.StatusOK, []byte(`{"card":"card","code":"f74d0ef2-f23b-46f9-9706-2815d2912c24","url":"https://pay"}`)), nil
	}))

	payment, err := c.CreatePayment(context.Background(), CreatePaymentOptions{RedirectURL: "https://discord.com"})
	require.NoError(t, err)
	assert.Equal(t, "card", payment.Card)
	assert.Equal(t, "f74d0ef2-f23b-46f9-9706-2815d2912c24", payment.Code)
}

func TestCreatePayment_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.CreatePayment(context.Background(), CreatePaymentOptions{RedirectURL: "https://discord.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreatePayment:")
	assert.Contains(t, err.Error(), "API error 400")
}

func TestCreateTransaction_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://api.test/transactions", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`{"balance":500}`)), nil
	}))

	resp, err := c.CreateTransaction(context.Background(), CreateTransactionOptions{Receiver: "1234", Amount: 10, Comment: "test"})
	require.NoError(t, err)
	assert.Equal(t, 500, resp.Balance)
}

func TestCreateTransaction_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.CreateTransaction(context.Background(), CreateTransactionOptions{Receiver: "1234", Amount: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateTransaction:")
	assert.Contains(t, err.Error(), "API error 400")
}

func TestUpdateWebhook_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "https://api.test/card/webhook", req.URL.String())
		return buildResponse(http.StatusOK, []byte(`{"id":"card-1","webhook":"https://webhook.com"}`)), nil
	}))

	resp, err := c.UpdateWebhook(context.Background(), UpdateWebhookOptions{URL: "https://webhook.com"})
	require.NoError(t, err)
	assert.Equal(t, "card-1", resp.ID)
	assert.Equal(t, "https://webhook.com", resp.Webhook)
}

func TestUpdateWebhook_ErrorWrapsEndpoint(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return buildResponse(http.StatusBadRequest, []byte(`{"message":"not found","error":"INVALID"}`)), nil
	}))

	_, err := c.UpdateWebhook(context.Background(), UpdateWebhookOptions{URL: "https://webhook.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateWebhook:")
	assert.Contains(t, err.Error(), "API error 400")
}
