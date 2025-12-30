package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/security/authentication/oidc"
)

func TestUser_HasGroupAccess(t *testing.T) {
	tests := []struct {
		name          string
		userGroups    []string
		allowedGroups []string
		expected      bool
	}{
		{
			name:          "user in allowed group",
			userGroups:    []string{"dev", "admin"},
			allowedGroups: []string{"admin"},
			expected:      true,
		},
		{
			name:          "user in first allowed group",
			userGroups:    []string{"admin", "dev"},
			allowedGroups: []string{"admin", "ops"},
			expected:      true,
		},
		{
			name:          "user in second allowed group",
			userGroups:    []string{"ops"},
			allowedGroups: []string{"admin", "ops"},
			expected:      true,
		},
		{
			name:          "user not in any allowed group",
			userGroups:    []string{"dev"},
			allowedGroups: []string{"admin", "ops"},
			expected:      false,
		},
		{
			name:          "empty allowed groups",
			userGroups:    []string{"dev", "admin"},
			allowedGroups: []string{},
			expected:      false,
		},
		{
			name:          "empty user groups",
			userGroups:    []string{},
			allowedGroups: []string{"admin"},
			expected:      false,
		},
		{
			name:          "both empty",
			userGroups:    []string{},
			allowedGroups: []string{},
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := oidc.User{
				Groups: tt.userGroups,
			}

			result := user.HasGroupAccess(tt.allowedGroups)

			assert.Equal(t, tt.expected, result, "should return expected access status")
		})
	}
}
