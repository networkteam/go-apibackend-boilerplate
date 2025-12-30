package oidc

import (
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/friendsofgo/errors"
)

// User represents a user with additional information in the OIDC authentication system.
type User struct {
	Subject       string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	// Groups as paths the user is a member of, either directly or through an ancestor group.
	Groups []string `json:"groups"`
}

func (u User) HasGroupAccess(allowedGroups []string) bool {
	for _, allowedGroup := range allowedGroups {
		if slices.Contains(u.Groups, allowedGroup) {
			return true
		}
	}
	return false
}

func userFromUserInfo(userInfo *oidc.UserInfo) (User, error) {
	user := User{
		Subject:       userInfo.Subject,
		Profile:       userInfo.Profile,
		Email:         userInfo.Email,
		EmailVerified: userInfo.EmailVerified,
	}

	if err := userInfo.Claims(&user); err != nil {
		return User{}, errors.Wrap(err, "getting user info claims")
	}

	return user, nil
}
