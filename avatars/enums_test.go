package avatars

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPart_String(t *testing.T) {
	assert.Equal(t, "head", Head.String())
	assert.Equal(t, "frontbust", Body.String())
	assert.Equal(t, "custom", Part("custom").String())
}

func TestPart_IsValid(t *testing.T) {
	tests := []struct {
		name string
		part Part
		want bool
	}{
		{"head is valid", Head, true},
		{"front is valid", Front, true},
		{"body is valid", Body, true},
		{"bust is valid", Bust, true},
		{"full is valid", Full, true},
		{"frontfull is valid", FrontFull, true},
		{"face is valid", Face, true},
		{"skin is valid", Skin, true},
		{"cape is valid", Cape, true},
		{"empty is invalid", Part(""), false},
		{"unknown is invalid", Part("portrait"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.part.IsValid())
		})
	}
}

func TestPart_IsTexture(t *testing.T) {
	tests := []struct {
		name string
		part Part
		want bool
	}{
		{"skin is texture", Skin, true},
		{"cape is texture", Cape, true},
		{"head is not texture", Head, false},
		{"bust is not texture", Bust, false},
		{"unknown is not texture", Part("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.part.IsTexture())
		})
	}
}

func TestPart_Is3D(t *testing.T) {
	tests := []struct {
		name string
		part Part
		want bool
	}{
		{"bust is 3D", Bust, true},
		{"full is 3D", Full, true},
		{"frontfull is not 3D", FrontFull, false},
		{"head is not 3D", Head, false},
		{"skin is not 3D", Skin, false},
		{"unknown is not 3D", Part("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.part.Is3D())
		})
	}
}

func TestParsePart(t *testing.T) {
	t.Run("valid parts", func(t *testing.T) {
		for _, s := range []string{"head", "front", "frontbust", "bust", "full", "frontfull", "face", "skin", "cape"} {
			p, err := ParsePart(s)
			require.NoError(t, err)
			assert.Equal(t, Part(s), p)
		}
	})

	t.Run("invalid part", func(t *testing.T) {
		p, err := ParsePart("portrait")
		require.Error(t, err)
		assert.Equal(t, Part(""), p)
		assert.Contains(t, err.Error(), "unknown avatar part")
		assert.Contains(t, err.Error(), "portrait")
	})

	t.Run("empty string is invalid", func(t *testing.T) {
		_, err := ParsePart("")
		require.Error(t, err)
	})
}

func TestSizeConstants(t *testing.T) {
	assert.Equal(t, 0, SizeDefault)
	assert.Equal(t, 64, Size64)
	assert.Equal(t, 128, Size128)
	assert.Equal(t, 256, Size256)
	assert.Equal(t, 512, Size512)
}
