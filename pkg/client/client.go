package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/certfix/certfix-cli/pkg/logger"
)

// HTTPClient represents an HTTP client for API requests
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Post makes a POST request
func (c *HTTPClient) Post(endpoint string, payload interface{}) (map[string]interface{}, error) {
	return c.request("POST", endpoint, payload, "")
}

// PostWithAuth makes a POST request with authentication
func (c *HTTPClient) PostWithAuth(endpoint string, payload interface{}, token string) (map[string]interface{}, error) {
	return c.request("POST", endpoint, payload, token)
}

// Get makes a GET request
func (c *HTTPClient) Get(endpoint string) (map[string]interface{}, error) {
	return c.request("GET", endpoint, nil, "")
}

// GetWithAuth makes a GET request with authentication
func (c *HTTPClient) GetWithAuth(endpoint string, token string) (map[string]interface{}, error) {
	return c.request("GET", endpoint, nil, token)
}

// PutWithAuth makes a PUT request with authentication
func (c *HTTPClient) PutWithAuth(endpoint string, payload interface{}, token string) (map[string]interface{}, error) {
	return c.request("PUT", endpoint, payload, token)
}

// DeleteWithAuth makes a DELETE request with authentication
func (c *HTTPClient) DeleteWithAuth(endpoint string, token string) (map[string]interface{}, error) {
	return c.request("DELETE", endpoint, nil, token)
}

// DeleteWithAuthAndPayload makes a DELETE request with authentication and payload
func (c *HTTPClient) DeleteWithAuthAndPayload(endpoint string, payload interface{}, token string) (map[string]interface{}, error) {
	return c.request("DELETE", endpoint, payload, token)
}

// PatchWithAuth makes a PATCH request with authentication
func (c *HTTPClient) PatchWithAuth(endpoint string, payload interface{}, token string) (map[string]interface{}, error) {
	return c.request("PATCH", endpoint, payload, token)
}

// request performs an HTTP request
func (c *HTTPClient) request(method, endpoint string, payload interface{}, token string) (map[string]interface{}, error) {
	log := logger.GetLogger()

	url := c.baseURL + endpoint
	log.Debugf("%s %s", method, url)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "certfix-cli/1.0")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Debugf("Response status: %d, body: %s", resp.StatusCode, string(responseBody))

		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return nil, fmt.Errorf("session expired or unauthorized: please run 'certfix login'")
		}

		// Extract message from standardized error response format
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil {
			// Check for details.message pattern (nested map)
			if details, ok := errorResponse["details"].(map[string]interface{}); ok {
				if message, ok := details["message"].(string); ok {
					return nil, fmt.Errorf("%s", message)
				}
			}
			// Check for top-level message field
			if message, ok := errorResponse["message"].(string); ok {
				return nil, fmt.Errorf("%s", message)
			}
			// Check for top-level error field
			if errMsg, ok := errorResponse["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}

		// Fallback to full error message
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response - handle objects, arrays, and non-JSON bodies
	var result map[string]interface{}
	if len(responseBody) > 0 {
		if responseBody[0] == '[' {
			// Response is an array, wrap it in an object
			var arrayResult []interface{}
			if err := json.Unmarshal(responseBody, &arrayResult); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
			result = map[string]interface{}{
				"_is_array":   true,
				"_array_data": arrayResult,
			}
		} else if responseBody[0] == '{' {
			// Response is an object
			if err := json.Unmarshal(responseBody, &result); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
		} else {
			// Non-JSON body (e.g. plain number or string) — treat as success with empty map
			result = map[string]interface{}{}
		}
	}

	return result, nil
}
