package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/networkteam/slogutils"
	"golang.org/x/oauth2"
)

// Middleware provides OIDC authentication middleware
type Middleware struct {
	config       *Config
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	provider     *oidc.Provider
	store        *sessions.CookieStore
}

// NewMiddleware creates a new OIDC middleware instance
func NewMiddleware(config *Config) (*Middleware, error) {
	// Set defaults and validate config
	config.SetDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize OIDC provider
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, config.GitLabURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})

	// Initialize session store
	store := sessions.NewCookieStore(config.SessionSecret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(config.SessionMaxAge.Seconds()),
		HttpOnly: config.SessionHTTPOnly,
		Secure:   config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	return &Middleware{
		config:       config,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		provider:     provider,
		store:        store,
	}, nil
}

// Wrap wraps an http.Handler with OIDC authentication
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc(m.config.LoginPath, m.handleLogin)
	mux.HandleFunc(m.config.CallbackPath, m.handleCallback)
	mux.HandleFunc(m.config.LogoutPath, m.handleLogout)

	// Protected handler
	mux.Handle("/", m.requireAuth(next))

	return mux
}

// requireAuth middleware that checks for valid authentication
func (m *Middleware) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slogutils.FromContext(ctx).
			With("middleware", "oidc")

		// Skip auth for auth endpoints
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get session
		session, err := m.store.Get(r, m.config.SessionName)
		if err != nil {
			logger.WarnContext(ctx, "Failed to get session", slogutils.Err(err))
			m.redirectToLogin(w, r)
			return
		}

		// Check if user is authenticated
		userSubjectInterface, sessionValueExists := session.Values["user"]
		if !sessionValueExists {
			m.redirectToLogin(w, r)
			return
		}

		userSubject, userSubjectOk := userSubjectInterface.(string)
		if !userSubjectOk || userSubject == "" {
			logger.WarnContext(ctx, "Invalid user subject in session")
			m.redirectToLogin(w, r)
			return
		}
		logger.DebugContext(ctx, "Found user in session", "user", userSubject)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// handleLogin initiates the OAuth2 flow
func (m *Middleware) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slogutils.FromContext(ctx).
		With("middleware", "oidc")

	// Generate state parameter
	state, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state in session
	session, _ := m.store.Get(r, m.config.SessionName) //nolint:errcheck
	if session == nil {
		session = sessions.NewSession(m.store, m.config.SessionName)
	}
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		logger.WarnContext(ctx, "Failed to save session", slogutils.Err(err))
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Redirect to GitLab
	authURL := m.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleCallback processes the OAuth2 callback
func (m *Middleware) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slogutils.FromContext(ctx).
		With("middleware", "oidc")

	// Verify state parameter
	session, err := m.store.Get(r, m.config.SessionName)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get session", slogutils.Err(err))
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	expectedState, sessionStateOk := session.Values["state"].(string)
	if !sessionStateOk || expectedState != r.URL.Query().Get("state") {
		logger.WarnContext(ctx, "Invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Handle OAuth2 errors
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		logger.WarnContext(ctx, "OAuth2 error", "error", errParam, "description", r.URL.Query().Get("error_description"))
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := m.oauth2Config.Exchange(ctx, code)
	if err != nil {
		logger.WarnContext(ctx, "Failed to exchange code for token", slogutils.Err(err))
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Extract and verify ID token
	rawIDToken, sessionStateOk := token.Extra("id_token").(string)
	if !sessionStateOk {
		http.Error(w, "Missing ID token", http.StatusInternalServerError)
		return
	}

	idToken, err := m.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.WarnContext(ctx, "Failed to verify ID token", slogutils.Err(err))
		http.Error(w, "Token verification failed", http.StatusInternalServerError)
		return
	}
	_ = idToken

	userInfo, err := m.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		logger.WarnContext(ctx, "Failed to get userinfo", slogutils.Err(err))
		http.Error(w, "Failed to get userinfo", http.StatusInternalServerError)
		return
	}

	user, err := userFromUserInfo(userInfo)
	if err != nil {
		logger.WarnContext(ctx, "Failed to parse userinfo", slogutils.Err(err))
		http.Error(w, "Failed to parse userinfo", http.StatusInternalServerError)
		return
	}

	if !user.HasGroupAccess(m.config.GitLabAllowedGroups) {
		logger.WarnContext(ctx, "User denied access - not in allowed groups", "subject", user.Subject, "groups", user.Groups, "allowedGroups", m.config.GitLabAllowedGroups)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	session.Values["user"] = user.Subject
	delete(session.Values, "state") // Clean up state
	if err := session.Save(r, w); err != nil {
		logger.WarnContext(ctx, "Failed to save session", slogutils.Err(err))
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Redirect to original destination or root
	returnTo := r.URL.Query().Get("return_to")
	if returnTo == "" {
		returnTo = m.config.PathPrefix + "/"
	}
	http.Redirect(w, r, returnTo, http.StatusFound)
}

// handleLogout clears the session and redirects
func (m *Middleware) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slogutils.FromContext(ctx).
		With("middleware", "oidc")

	session, _ := m.store.Get(r, m.config.SessionName) //nolint:errcheck
	if session == nil {
		session = sessions.NewSession(m.store, m.config.SessionName)
	}

	// Clear session
	session.Values = make(map[any]any)
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		logger.WarnContext(ctx, "Failed to clear session", slogutils.Err(err))
	}

	// Redirect to login
	http.Redirect(w, r, m.config.PathPrefix+m.config.LoginPath, http.StatusFound)
}

// redirectToLogin redirects to the login page with return URL
func (m *Middleware) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	returnTo := url.QueryEscape(r.URL.String())
	loginURL := fmt.Sprintf("%s?return_to=%s", m.config.LoginPath, returnTo)
	http.Redirect(w, r, m.config.PathPrefix+loginURL, http.StatusFound)
}

// generateRandomString creates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
