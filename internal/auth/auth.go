package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const (
	apiKeyPrefix = "krane_"
)

type Credentials struct {
	Server string
	APIKey string
}

// ValidateAPIKey checks if the API key has the correct format
func ValidateAPIKey(apiKey string) error {
	if !strings.HasPrefix(apiKey, apiKeyPrefix) {
		return fmt.Errorf("invalid API key format: must start with %s", apiKeyPrefix)
	}
	if len(apiKey) < 16 {
		return fmt.Errorf("invalid API key format: too short")
	}
	return nil
}

// GetCredentialsFromEnv retrieves credentials from environment variables
func GetCredentialsFromEnv() (*Credentials, error) {
	server := os.Getenv("KRANE_SERVER")
	apiKey := os.Getenv("KRANE_API_KEY")

	if server == "" {
		server = "http://localhost:8080"
	}

	if apiKey == "" {
		return nil, fmt.Errorf("KRANE_API_KEY not set")
	}

	return &Credentials{
		Server: server,
		APIKey: apiKey,
	}, nil
}

// GetAuthHeader returns the Authorization header value
func GetAuthHeader(apiKey string) string {
	return fmt.Sprintf("Bearer %s", apiKey)
}

// SanitizeAPIKey returns a safe representation of the API key for logging
func SanitizeAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:8] + "***"
}

// CompareAPIKeys securely compares two API keys
func CompareAPIKeys(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// EncodeBase64 encodes a string to base64
func EncodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// DecodeBase64 decodes a base64 string
func DecodeBase64(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
