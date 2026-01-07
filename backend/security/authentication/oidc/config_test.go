package oidc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/security/authentication/oidc"
)

func TestConfig_SetDefaults(t *testing.T) {
	t.Run("sets all defaults for empty config", func(t *testing.T) {
		cfg := &oidc.Config{}
		cfg.SetDefaults()

		assert.Equal(t, "devlog_session", cfg.SessionName, "should set default session name")
		assert.Equal(t, 24*time.Hour, cfg.SessionMaxAge, "should set default session max age")
		assert.Equal(t, "/auth/login", cfg.LoginPath, "should set default login path")
		assert.Equal(t, "/auth/callback", cfg.CallbackPath, "should set default callback path")
		assert.Equal(t, "/auth/logout", cfg.LogoutPath, "should set default logout path")
		assert.True(t, cfg.SessionHTTPOnly, "should set HTTP-only by default")
	})

	t.Run("preserves custom values", func(t *testing.T) {
		cfg := &oidc.Config{
			SessionName:   "custom_session",
			SessionMaxAge: 1 * time.Hour,
			LoginPath:     "/custom/login",
			CallbackPath:  "/custom/callback",
			LogoutPath:    "/custom/logout",
		}
		cfg.SetDefaults()

		assert.Equal(t, "custom_session", cfg.SessionName, "should preserve custom session name")
		assert.Equal(t, 1*time.Hour, cfg.SessionMaxAge, "should preserve custom session max age")
		assert.Equal(t, "/custom/login", cfg.LoginPath, "should preserve custom login path")
		assert.Equal(t, "/custom/callback", cfg.CallbackPath, "should preserve custom callback path")
		assert.Equal(t, "/custom/logout", cfg.LogoutPath, "should preserve custom logout path")
	})
}

func TestConfig_Validate(t *testing.T) {
	validConfig := func() *oidc.Config {
		return &oidc.Config{
			GitLabURL:     "https://gitlab.example.com",
			ClientID:      "client-id",
			ClientSecret:  "client-secret",
			RedirectURL:   "https://app.example.com/auth/callback",
			SessionSecret: []byte("session-secret"),
		}
	}

	tests := []struct {
		name        string
		modifyConfig func(*oidc.Config)
		expectedErr error
	}{
		{
			name:         "valid config",
			modifyConfig: func(c *oidc.Config) {},
			expectedErr:  nil,
		},
		{
			name:         "missing GitLabURL",
			modifyConfig: func(c *oidc.Config) { c.GitLabURL = "" },
			expectedErr:  oidc.ErrMissingGitLabURL,
		},
		{
			name:         "missing ClientID",
			modifyConfig: func(c *oidc.Config) { c.ClientID = "" },
			expectedErr:  oidc.ErrMissingClientID,
		},
		{
			name:         "missing ClientSecret",
			modifyConfig: func(c *oidc.Config) { c.ClientSecret = "" },
			expectedErr:  oidc.ErrMissingClientSecret,
		},
		{
			name:         "missing RedirectURL",
			modifyConfig: func(c *oidc.Config) { c.RedirectURL = "" },
			expectedErr:  oidc.ErrMissingRedirectURL,
		},
		{
			name:         "missing SessionSecret",
			modifyConfig: func(c *oidc.Config) { c.SessionSecret = nil },
			expectedErr:  oidc.ErrMissingSessionSecret,
		},
		{
			name:         "empty SessionSecret",
			modifyConfig: func(c *oidc.Config) { c.SessionSecret = []byte{} },
			expectedErr:  oidc.ErrMissingSessionSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modifyConfig(cfg)

			err := cfg.Validate()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr, "should return expected error")
			} else {
				require.NoError(t, err, "should not return error")
			}
		})
	}
}
