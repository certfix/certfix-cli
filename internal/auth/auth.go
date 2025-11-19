package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// TokenData represents the stored authentication token
type TokenData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login authenticates using a personal access token and returns a JWT token
func Login(email, personalToken, endpoint string) (string, error) {
	log := logger.GetLogger()

	if endpoint == "" {
		endpoint = config.GetDefaultEndpoint()
	}

	log.Debugf("Authenticating with personal token at endpoint: %s", endpoint)

	// Create API client
	apiClient := client.NewHTTPClient(endpoint)

	// Perform CLI auth request
	payload := map[string]string{
		"email":                email,
		"personal_access_token": personalToken,
	}

	response, err := apiClient.Post("/auth/cli", payload)
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %w", err)
	}

	// Extract token from response
	token, ok := response["token"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response: token not found")
	}

	return token, nil
}

// StoreToken saves the authentication token to disk
func StoreToken(token string) error {
	log := logger.GetLogger()

	// Parse token to get expiration
	claims := jwt.MapClaims{}
	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(token, claims)
	if err != nil {
		log.WithError(err).Warn("Failed to parse token claims, using default expiration")
	}

	expiresAt := time.Now().Add(24 * time.Hour) // Default 24 hours
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	}

	tokenData := TokenData{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	// Get token file path
	tokenPath := getTokenPath()

	// Create directory if it doesn't exist
	tokenDir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Marshal token data
	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Write token to file
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	log.Debugf("Token stored at: %s", tokenPath)
	return nil
}

// GetToken retrieves the stored authentication token
func GetToken() (string, error) {
	tokenPath := getTokenPath()

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not authenticated: please run 'certfix login'")
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return "", fmt.Errorf("failed to parse token file: %w", err)
	}

	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		return "", fmt.Errorf("token expired: please run 'certfix login'")
	}

	return tokenData.Token, nil
}

// IsAuthenticated checks if the user is currently authenticated
func IsAuthenticated() bool {
	_, err := GetToken()
	return err == nil
}

// Logout removes the stored authentication token
func Logout() error {
	tokenPath := getTokenPath()

	if err := os.Remove(tokenPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already logged out
		}
		return fmt.Errorf("failed to remove token file: %w", err)
	}

	return nil
}

// getTokenPath returns the path to the token file
func getTokenPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".certfix", "token.json")
}
