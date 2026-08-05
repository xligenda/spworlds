package avatars

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSize(t *testing.T) {
	assert.Equal(t, "0", Size(SizeDefault))
	assert.Equal(t, "64", Size(64))
	assert.Equal(t, "960", Size(960))
	assert.Equal(t, "-1", Size(-1))
}

func TestClient_URL(t *testing.T) {
	t.Run("builds expected URL", func(t *testing.T) {
		c := NewClient(WithBaseURL("https://avatars.test"))
		got := c.URL("5opka", Head, 128)
		assert.Equal(t, "https://avatars.test/head/5opka?width=128", got)
	})

	t.Run("trims trailing slash from base URL", func(t *testing.T) {
		c := NewClient(WithBaseURL("https://avatars.test/"))
		got := c.URL("5opka", Bust, 64)
		assert.Equal(t, "https://avatars.test/bust/5opka?width=64", got)
	})

	t.Run("escapes player name", func(t *testing.T) {
		c := NewClient(WithBaseURL("https://avatars.test"))
		got := c.URL("5opka/escape", Skin, 0)
		assert.Equal(t, "https://avatars.test/skin/5opka%2Fescape", got)
	})

	t.Run("negative width falls back to default size", func(t *testing.T) {
		c := NewClient(WithBaseURL("https://avatars.test"))
		got := c.URL("5opka", Cape, -50)
		assert.Equal(t, "https://avatars.test/cape/5opka", got)
	})

	t.Run("default base URL is used when not overridden", func(t *testing.T) {
		c := NewClient()
		got := c.URL("5opka", Face, 256)
		assert.Equal(t, DefaultBaseURL+"/face/5opka?width=256", got)
	})
}

func TestClient_ParsedURL(t *testing.T) {
	t.Run("parses valid URL", func(t *testing.T) {
		c := NewClient(WithBaseURL("https://avatars.test"))
		u, err := c.ParsedURL("b963413a-b97f-4124-aebf-9a1eefd0b144", Full, 512)
		require.NoError(t, err)
		require.NotNil(t, u)
		assert.Equal(t, "https", u.Scheme)
		assert.Equal(t, "avatars.test", u.Host)
		assert.Equal(t, "/full/b963413a-b97f-4124-aebf-9a1eefd0b144", u.Path)
		assert.Equal(t, "width=512", u.RawQuery)
	})

	t.Run("propagates parse error", func(t *testing.T) {
		c := NewClient(WithBaseURL("://bad-url"))
		_, err := c.ParsedURL("5opka", Head, 0)
		require.Error(t, err)
	})
}
