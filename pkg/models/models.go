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

// CertfixConfig represents the complete YAML configuration file
type CertfixConfig struct {
	Events        []EventConfig        `yaml:"events"`
	Policies      []PolicyConfig       `yaml:"policies"`
	ServiceGroups []ServiceGroupConfig `yaml:"service_groups"`
	Services      []ServiceConfig      `yaml:"services"`
}

// EventConfig represents an event configuration
type EventConfig struct {
	Name     string `yaml:"name"`
	Severity string `yaml:"severity"`
	Enabled  bool   `yaml:"enabled"`
}

// PolicyConfig represents a policy configuration
type PolicyConfig struct {
	Name     string                 `yaml:"name"`
	Strategy string                 `yaml:"strategy"`
	Enabled  bool                   `yaml:"enabled"`
	CronConfig map[string]string    `yaml:"cron_config,omitempty"`
	EventConfig map[string]interface{} `yaml:"event_config,omitempty"`
}

// ServiceGroupConfig represents a service group configuration
type ServiceGroupConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Enabled     bool   `yaml:"enabled"`
}

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Hash        string                  `yaml:"hash"`
	Name        string                  `yaml:"name"`
	Active      bool                    `yaml:"active"`
	WebhookURL  string                  `yaml:"webhook_url,omitempty"`
	GroupName   string                  `yaml:"group_name,omitempty"`    // Reference by name
	PolicyName  string                  `yaml:"policy_name,omitempty"`   // Reference by name
	Keys        []ServiceKeyConfig      `yaml:"keys,omitempty"`
	Relations   []ServiceRelationConfig `yaml:"relations,omitempty"`
}

// ServiceKeyConfig represents an API key configuration
type ServiceKeyConfig struct {
	Name           string `yaml:"name"`
	Enabled        bool   `yaml:"enabled"`
	ExpirationDays int    `yaml:"expiration_days,omitempty"`
}

// ServiceRelationConfig represents a service relation (matriz)
type ServiceRelationConfig struct {
	TargetHash string `yaml:"target_hash"`
	Type       string `yaml:"type,omitempty"`
}

// CreatedResource tracks resources created during apply for rollback
type CreatedResource struct {
	Type string // "evento", "politica", "service_group", "service", "key", "relation"
	Hash string // Primary identifier (hash or ID)
	ID   string // Secondary identifier (for keys and relations)
}
