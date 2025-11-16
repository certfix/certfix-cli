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

// ListCertificates lists all certificates
func (c *Client) ListCertificates() ([]*models.Certificate, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.GetWithAuth("/certificates", token)
	if err != nil {
		return nil, err
	}

	// Parse response into Certificate models
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
