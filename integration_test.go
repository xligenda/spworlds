package spworlds_test

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xligenda/spworlds"
)

func TestIntegration_Me(t *testing.T) {
	client := setupIntegrationClient(t)

	user, err := client.Me(context.Background())

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotNil(t, user.Username)
}

func TestIntegration_ClientCard(t *testing.T) {
	client := setupIntegrationClient(t)

	card, err := client.ClientCard(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, card)
}

func TestIntegration_SetWebhook(t *testing.T) {
	client := setupIntegrationClient(t)

	webhookURL := "https://example.com/webhook"
	resp, err := client.UpdateWebhook(context.Background(), spworlds.UpdateWebhookOptions{URL: webhookURL})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, webhookURL, resp.Webhook)

	card, err := client.ClientCard(context.Background())

	require.NoError(t, err)
	require.NotNil(t, card)
	require.NotNil(t, card.Webhook)
	assert.Equal(t, webhookURL, *card.Webhook)
}

func setupIntegrationClient(t *testing.T) *spworlds.Client {
	t.Helper()
	id, token, ok := integrationCredentials()
	if !ok {
		t.Skip("Skipping integration test: SPWORLDS_API_ID and SPWORLDS_API_TOKEN (or SPWORLDS_API_KEY) are not set")
	}
	return spworlds.NewClient(id, token, nil)
}

func integrationCredentials() (string, string, bool) {
	id := strings.TrimSpace(os.Getenv("SPWORLDS_API_ID"))
	token := strings.TrimSpace(os.Getenv("SPWORLDS_API_TOKEN"))
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
