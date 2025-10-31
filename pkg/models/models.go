package models

// Instance represents a Certfix instance
type Instance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Region string `json:"region"`
	Status string `json:"status"`
}

// Certificate represents an SSL/TLS certificate
type Certificate struct {
	ID        string `json:"id"`
	Domain    string `json:"domain"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

// User represents a user account
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	User      User   `json:"user"`
}
