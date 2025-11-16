package api

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/certfix/certfix-cli/pkg/models"
)

// Client represents an API client
type Client struct {
	httpClient *client.HTTPClient
}

// NewClient creates a new API client
func NewClient() *Client {
	endpoint := config.GetDefaultEndpoint()
	return &Client{
		httpClient: client.NewHTTPClient(endpoint),
	}
}

// CreateInstance creates a new instance
func (c *Client) CreateInstance(name, instanceType, region string) (*models.Instance, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]string{
		"name":   name,
		"type":   instanceType,
		"region": region,
	}

	response, err := c.httpClient.PostWithAuth("/instances", payload, token)
	if err != nil {
		return nil, err
	}

	// Parse response into Instance model
	instance := &models.Instance{
		ID:     fmt.Sprintf("%v", response["id"]),
		Name:   name,
		Status: fmt.Sprintf("%v", response["status"]),
	}

	return instance, nil
}

// ListInstances lists all instances
func (c *Client) ListInstances() ([]*models.Instance, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.GetWithAuth("/instances", token)
	if err != nil {
		return nil, err
	}

	// Parse response into Instance models
	instances := []*models.Instance{}
	if items, ok := response["instances"].([]interface{}); ok {
		for _, item := range items {
			if inst, ok := item.(map[string]interface{}); ok {
				instance := &models.Instance{
					ID:     fmt.Sprintf("%v", inst["id"]),
					Name:   fmt.Sprintf("%v", inst["name"]),
					Status: fmt.Sprintf("%v", inst["status"]),
				}
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// DeleteInstance deletes an instance
func (c *Client) DeleteInstance(id string) error {
	token, err := auth.GetToken()
	if err != nil {
		return err
	}

	_, err = c.httpClient.DeleteWithAuth(fmt.Sprintf("/instances/%s", id), token)
	return err
}

// CreateCertificate creates a new certificate
func (c *Client) CreateCertificate(commonName, certType, description string, days, keySize int, san string) (map[string]interface{}, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	// Build payload with required fields
	payload := map[string]interface{}{
		"commonName": commonName,
		"type":       certType,
	}

	// Add optional fields only if provided
	if description != "" {
		payload["description"] = description
	}
	if days > 0 {
		payload["days"] = days
	}
	if keySize > 0 {
		payload["keySize"] = keySize
	}
	if san != "" {
		payload["san"] = san
	}

	response, err := c.httpClient.PostWithAuth("/certificates", payload, token)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ListValidCertificates lists all valid certificates
func (c *Client) ListValidCertificates() ([]map[string]interface{}, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.GetWithAuth("/certificates", token)
	if err != nil {
		return nil, err
	}

	// Check if response is wrapped in a "certificates" key or is a direct array
	if certs, ok := response["certificates"].([]interface{}); ok {
		return convertToMapArray(certs), nil
	}
	
	// If response has an "array" marker indicating direct array response
	if response["_is_array"] != nil {
		if arr, ok := response["_array_data"].([]interface{}); ok {
			return convertToMapArray(arr), nil
		}
	}

	return []map[string]interface{}{}, nil
}

// ListRevokedCertificates lists all revoked certificates
func (c *Client) ListRevokedCertificates() ([]map[string]interface{}, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.GetWithAuth("/certificates/revoked", token)
	if err != nil {
		return nil, err
	}

	// Check if response is wrapped in a "certificates" key or is a direct array
	if certs, ok := response["certificates"].([]interface{}); ok {
		return convertToMapArray(certs), nil
	}
	
	// If response has an "array" marker indicating direct array response
	if response["_is_array"] != nil {
		if arr, ok := response["_array_data"].([]interface{}); ok {
			return convertToMapArray(arr), nil
		}
	}

	return []map[string]interface{}{}, nil
}

// ListExpiringCertificates lists certificates expiring in the specified number of days
func (c *Client) ListExpiringCertificates(days string) ([]map[string]interface{}, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/certificates?expiringInDays=%s", days)
	response, err := c.httpClient.GetWithAuth(endpoint, token)
	if err != nil {
		return nil, err
	}

	// Check if response is wrapped in a "certificates" key or is a direct array
	if certs, ok := response["certificates"].([]interface{}); ok {
		return convertToMapArray(certs), nil
	}
	
	// If response has an "array" marker indicating direct array response
	if response["_is_array"] != nil {
		if arr, ok := response["_array_data"].([]interface{}); ok {
			return convertToMapArray(arr), nil
		}
	}

	return []map[string]interface{}{}, nil
}

// convertToMapArray converts []interface{} to []map[string]interface{}
func convertToMapArray(items []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if cert, ok := item.(map[string]interface{}); ok {
			result = append(result, cert)
		}
	}
	return result
}

// parseCertificatesList is a helper function to parse certificate list responses (deprecated)
func parseCertificatesList(response map[string]interface{}) ([]*models.Certificate, error) {
	certificates := []*models.Certificate{}
	if items, ok := response["certificates"].([]interface{}); ok {
		for _, item := range items {
			if cert, ok := item.(map[string]interface{}); ok {
				certificate := &models.Certificate{
					ID:        fmt.Sprintf("%v", cert["id"]),
					Domain:    fmt.Sprintf("%v", cert["domain"]),
					Status:    fmt.Sprintf("%v", cert["status"]),
					ExpiresAt: fmt.Sprintf("%v", cert["expires_at"]),
				}
				certificates = append(certificates, certificate)
			}
		}
	}

	return certificates, nil
}

// ListCertificates lists all certificates (deprecated - kept for compatibility)
func (c *Client) ListCertificates() ([]*models.Certificate, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.GetWithAuth("/certificates", token)
	if err != nil {
		return nil, err
	}

	return parseCertificatesList(response)
}

// RenewCertificate renews a certificate
func (c *Client) RenewCertificate(id string) (*models.Certificate, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.PostWithAuth(fmt.Sprintf("/certificates/%s/renew", id), nil, token)
	if err != nil {
		return nil, err
	}

	// Parse response into Certificate model
	cert := &models.Certificate{
		ID:        id,
		Domain:    fmt.Sprintf("%v", response["domain"]),
		Status:    fmt.Sprintf("%v", response["status"]),
		ExpiresAt: fmt.Sprintf("%v", response["expires_at"]),
	}

	return cert, nil
}

// RevokeCertificate revokes a certificate
func (c *Client) RevokeCertificate(id string) error {
	token, err := auth.GetToken()
	if err != nil {
		return err
	}

	_, err = c.httpClient.DeleteWithAuth(fmt.Sprintf("/certificates/%s", id), token)
	return err
}

// CreateBackup creates a backup of the Certificate Authority
func (c *Client) CreateBackup() (map[string]interface{}, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.PostWithAuth("/ca/backup", nil, token)
	if err != nil {
		return nil, err
	}

	return response, nil
}
