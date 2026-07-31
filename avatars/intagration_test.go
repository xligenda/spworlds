package avatars_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xligenda/spworlds/avatars"
)

// 8-byte signature every PNG file starts with
var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

func TestIntegration_Head(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	data, err := client.Head(ctx, integrationPlayer(), avatars.Size128)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.True(t, bytes.HasPrefix(data, pngSignature), "response does not look like a PNG image")
}

func TestIntegration_AllParts(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	player := integrationPlayer()

	parts := []avatars.Part{
		avatars.Head,
		avatars.Front,
		avatars.Body,
		avatars.Bust,
		avatars.Full,
		avatars.FrontFull,
		avatars.Face,
		avatars.Skin,
	}

	for _, part := range parts {
		part := part
		t.Run(part.String(), func(t *testing.T) {
			t.Parallel()
			ctx := newTestContext(t)

			data, err := client.Fetch(ctx, player, part, avatars.SizeDefault)
			require.NoError(t, err)
			assert.NotEmpty(t, data)
			assert.True(t, bytes.HasPrefix(data, pngSignature), "response does not look like a PNG image")
		})
	}
}

func TestIntegration_FetchTo(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	var buf bytes.Buffer
	n, err := client.FetchTo(ctx, &buf, integrationPlayer(), avatars.Skin, avatars.Size64)
	require.NoError(t, err)
	assert.Positive(t, n)
	assert.Equal(t, int(n), buf.Len())
	assert.True(t, bytes.HasPrefix(buf.Bytes(), pngSignature), "response does not look like a PNG image")
}

func TestIntegration_UnknownPlayer(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	_, err := client.Head(ctx, "this-player-should-not-exist-00000000", avatars.SizeDefault)
	require.Error(t, err)
}

func TestIntegration_URLIsReachable(t *testing.T) {
	t.Parallel()
	client := setupIntegrationClient(t)
	ctx := newTestContext(t)

	u, err := client.ParsedURL(integrationPlayer(), avatars.Head, avatars.Size64)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "https", u.Scheme)

	data, err := client.Fetch(ctx, integrationPlayer(), avatars.Head, avatars.Size64)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func setupIntegrationClient(t *testing.T) *avatars.Client {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in -short mode")
	}

	return avatars.NewClient()
}

func integrationPlayer() string {
	return envOr("SPWORLDS_TEST_AVATARS_PLAYER", "xligenda")
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
