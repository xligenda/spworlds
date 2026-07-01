package spworlds

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("id", "token", nil)

	expectedKey := base64.StdEncoding.EncodeToString([]byte("id:token"))

	assert.Equal(t, DefaultAPIURL, c.apiURL)
	assert.Equal(t, expectedKey, c.apiKey)
	assert.NotNil(t, c.httpClient)
}

func TestNewClient_ConfigOverride(t *testing.T) {
	customClient := &http.Client{Timeout: 1 * time.Second}
	cfg := &ClientConfig{
		APIURL:     "https://example.com/api",
		UserAgent:  "test-agent/0.1",
		HTTPClient: customClient,
	}

	c := NewClient("id", "token", cfg)

	assert.Equal(t, cfg.APIURL, c.apiURL)
	assert.Equal(t, cfg.UserAgent, c.userAgent)
	assert.Equal(t, customClient, c.httpClient)
}

func TestRESTError_ErrorString(t *testing.T) {
	err := &RESTError{StatusCode: 401, ErrorCode: "UNAUTH", Message: "unauthorized"}
	errStr := err.Error()

	assert.Contains(t, errStr, "401")
	assert.Contains(t, errStr, "UNAUTH")
	assert.Contains(t, errStr, "unauthorized")
}

func TestTimestampUnix(t *testing.T) {
	t.Run("valid timestamp", func(t *testing.T) {
		tm := Timestamp("2026-07-01T12:00:00.000Z")
		expected := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC).Unix()
		assert.Equal(t, expected, tm.Unix())
	})

	t.Run("invalid timestamp fallback", func(t *testing.T) {
		assert.Zero(t, Timestamp("invalid").Unix())
	})
}

func TestCityMemberRole_IsMayor(t *testing.T) {
	assert.True(t, Mayor.IsMayor(), "Mayor should be recognized as mayor")
	assert.False(t, Member.IsMayor(), "Member should not be recognized as mayor")
}

func TestCityMember_IsMayor(t *testing.T) {
	assert.True(t, CityMember{Role: Mayor}.IsMayor())
	assert.False(t, CityMember{Role: DeputyMayor}.IsMayor())
	assert.False(t, CityMember{Role: Member}.IsMayor())
}
