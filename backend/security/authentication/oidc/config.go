package oidc

import (
	"time"
)

// Config holds the OIDC configuration
type Config struct {
	// GitLab instance URL (e.g., "https://gitlab.com" or "https://gitlab.example.com")
	GitLabURL string

	// OAuth2 client credentials from GitLab application registration
	ClientID     string
	ClientSecret string

	// Redirect URL for OAuth2 callback (must match GitLab app config)
	RedirectURL string

	// Optional: Restrict access to specific GitLab groups
	// If empty, any authenticated GitLab user is allowed
	GitLabAllowedGroups []string

	// Session configuration
	SessionSecret   []byte        // Secret for signing session cookies
	SessionName     string        // Cookie name (default: "devlog_session")
	SessionMaxAge   time.Duration // Session expiration (default: 24h)
	SessionSecure   bool          // Set secure flag on cookies (use true for HTTPS)
	SessionHTTPOnly bool          // Set HttpOnly flag on cookies (default: true)

	PathPrefix string

	// Optional: Custom login/logout paths
	LoginPath    string // Default: "/auth/login"
	CallbackPath string // Default: "/auth/callback"
	LogoutPath   string // Default: "/auth/logout"
}

// SetDefaults fills in default values for optional config fields
func (c *Config) SetDefaults() {
	if c.SessionName == "" {
		c.SessionName = "devlog_session"
	}
	if c.SessionMaxAge == 0 {
		c.SessionMaxAge = 24 * time.Hour
	}
	if c.LoginPath == "" {
		c.LoginPath = "/auth/login"
	}
	if c.CallbackPath == "" {
		c.CallbackPath = "/auth/callback"
	}
	if c.LogoutPath == "" {
		c.LogoutPath = "/auth/logout"
	}
	// Default to HttpOnly for security
	c.SessionHTTPOnly = true
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.GitLabURL == "" {
		return ErrMissingGitLabURL
	}
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	if c.RedirectURL == "" {
		return ErrMissingRedirectURL
	}
	if len(c.SessionSecret) == 0 {
		return ErrMissingSessionSecret
	}
	return nil
}
