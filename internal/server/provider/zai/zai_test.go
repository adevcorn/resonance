package zai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config with default base URL",
			config: Config{
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid config with custom base URL",
			config: Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.z.ai/v1",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: Config{
				BaseURL: "https://api.z.ai/v1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				require.NoError(t, err)
				require.NotNil(t, provider)
				assert.Equal(t, "zai", provider.Name())
				assert.True(t, provider.SupportsTools())
				assert.Equal(t, tt.config.APIKey, provider.apiKey)

				// Check base URL
				expectedBaseURL := tt.config.BaseURL
				if expectedBaseURL == "" {
					expectedBaseURL = DefaultBaseURL
				}
				assert.Equal(t, expectedBaseURL, provider.baseURL)
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	provider, err := NewProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)
	assert.Equal(t, "zai", provider.Name())
}

func TestProviderSupportsTools(t *testing.T) {
	provider, err := NewProvider(Config{APIKey: "test-key"})
	require.NoError(t, err)
	assert.True(t, provider.SupportsTools())
}
