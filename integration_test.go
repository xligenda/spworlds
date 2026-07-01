package spworlds_test

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xligenda/spworlds"
)

func TestIntegration_Me(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	user, err := client.Me(ctx)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.NotNil(t, user.Cards)
	assert.NotNil(t, user.Cities)
	assert.NotNil(t, user.CreatedAt)
	assert.NotNil(t, user.ID)
	assert.NotNil(t, user.Roles)
	assert.NotNil(t, user.Status)
	assert.NotNil(t, user.Username)
	assert.Nil(t, user.UUID)
}

func TestIntegration_ClientCard(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	card, err := client.ClientCard(ctx)
	require.NoError(t, err)
	assert.NotNil(t, card)
}

func TestIntegration_User(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	discordID := envOr("SPWORLDS_TEST_DISCORD_ID", "888016163844534372")
	wantUUID := envOr("SPWORLDS_TEST_UUID", "bec8a88f04fc4b56aa5bec1d3e0ccb66")
	wantUsername := envOr("SPWORLDS_TEST_USERNAME", "xligenda")

	user, err := client.User(ctx, discordID)
	require.NoError(t, err)
	require.NotNil(t, user)

	require.NotNil(t, user.UUID)
	require.NotNil(t, user.Username)
	assert.Equal(t, wantUUID, *user.UUID)
	assert.Equal(t, wantUsername, *user.Username)

	assert.Nil(t, user.Cards)
	assert.Nil(t, user.Cities)
	assert.Nil(t, user.CreatedAt)
	assert.Nil(t, user.ID)
	assert.Nil(t, user.Roles)
	assert.Nil(t, user.Status)
}

func TestIntegration_SetWebhook(t *testing.T) {
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	original, err := client.ClientCard(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		restoreCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		originalURL := ""
		if original.Webhook != nil {
			originalURL = *original.Webhook
		}
		_, _ = client.UpdateWebhook(restoreCtx, spworlds.UpdateWebhookOptions{URL: originalURL})
	})

	const webhookURL = "https://example.com/webhook"
	resp, err := client.UpdateWebhook(ctx, spworlds.UpdateWebhookOptions{URL: webhookURL})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, webhookURL, resp.Webhook)

	card, err := client.ClientCard(ctx)
	require.NoError(t, err)
	require.NotNil(t, card)
	require.NotNil(t, card.Webhook)
	assert.Equal(t, webhookURL, *card.Webhook)
}

func TestIntegration_CreateTransaction(t *testing.T) {
	if strings.TrimSpace(os.Getenv("SPWORLDS_TEST_ALLOW_POST")) != "1" {
		t.Skip("Skipping transaction test: set SPWORLDS_TEST_ALLOW_POST=1 to allow a balance transfer")
	}

	receiver := strings.TrimSpace(os.Getenv("SPWORLDS_TEST_RECEIVER_CARD"))
	if receiver == "" {
		t.Skip("Skipping transaction test: SPWORLDS_TEST_RECEIVER_CARD is not set")
	}

	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	card, err := client.ClientCard(ctx)
	require.NoError(t, err)
	require.NotNil(t, card)

	const amount = 1
	if card.Balance < amount {
		t.Skip("Skipping transaction test: insufficient balance")
	}

	tc, err := client.CreateTransaction(ctx, spworlds.CreateTransactionOptions{
		Receiver: receiver,
		Amount:   amount,
		Comment:  "Test transaction",
	})
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Equal(t, card.Balance-amount, tc.Balance)
}

func setupIntegrationClient(t *testing.T) *spworlds.Client {
	t.Helper()
	id, token, ok := integrationCredentials()
	if !ok {
		t.Skip("Skipping integration test: SPWORLDS_API_ID and SPWORLDS_API_TOKEN (or SPWORLDS_API_KEY) are not set")
	}
	return spworlds.NewClient(id, token, nil)
}

func integrationCredentials() (id, token string, ok bool) {
	id = strings.TrimSpace(os.Getenv("SPWORLDS_API_ID"))
	token = strings.TrimSpace(os.Getenv("SPWORLDS_API_TOKEN"))
	if id != "" && token != "" {
		return id, token, true
	}

	encoded := strings.TrimSpace(os.Getenv("SPWORLDS_API_KEY"))
	if encoded == "" {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func newTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
