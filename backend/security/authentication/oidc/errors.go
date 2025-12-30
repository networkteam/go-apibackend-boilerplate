package oidc

import "errors"

var (
	// Configuration errors
	ErrMissingGitLabURL     = errors.New("GitLab URL is required")
	ErrMissingClientID      = errors.New("OAuth2 client ID is required")
	ErrMissingClientSecret  = errors.New("OAuth2 client secret is required")
	ErrMissingRedirectURL   = errors.New("OAuth2 redirect URL is required")
	ErrMissingSessionSecret = errors.New("session secret is required")

	// Authentication errors
	ErrInvalidState      = errors.New("invalid OAuth2 state parameter")
	ErrTokenExchange     = errors.New("failed to exchange authorization code for token")
	ErrUserInfoRetrieval = errors.New("failed to retrieve user information")
	ErrAccessDenied      = errors.New("access denied: user not in allowed groups/projects")
	ErrInvalidSession    = errors.New("invalid or expired session")
)
