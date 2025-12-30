package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

//nolint:gochecknoglobals
var (
	fixedSystemAdminEmailAddressAccountIDMap = map[string]string{
		"admin@example.com": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
	}

	fixedOrganisationAdminEmailAddressAccountIDMap = map[string]string{
		"admin+acmeinc@example.com":      "3ad082c7-cbda-49e1-a707-c53e1962be65",
		"otheradmin+acmeinc@example.com": "f045e5d1-cdad-4964-a7e2-139c8a87346c",

		"admin+othercorp@example.com": "2035f4da-f385-42c4-a609-02d9aa7290e5",
	}

	fixedAccountOrganisationIDMap = map[string]string{
		"3ad082c7-cbda-49e1-a707-c53e1962be65": "6330de58-2761-411e-a243-bec6d0c53876", // Acme Inc.
		"f045e5d1-cdad-4964-a7e2-139c8a87346c": "6330de58-2761-411e-a243-bec6d0c53876", // Acme Inc.

		"2035f4da-f385-42c4-a609-02d9aa7290e5": "dba20d09-a3df-4975-9406-2fb6fd8f0940", // Other Corp.
	}

	FixedJWTSecret = "d8f42e3021b6163fc406983ccd4bf465bd15fbcab64d5e251c646f2da6e9dad9" //nolint:gosec

	FixedConfig = domain.Config{
		JWTSecret: FixedJWTSecret,
	}
)

type ApplyAuthValuesFunc func(t *testing.T, deps api.ResolverDependencies, req *http.Request) FixedAccessTokenData

func ApplyAuthValuesFuncSystemAdministrator(emailAddress string) ApplyAuthValuesFunc {
	return func(t *testing.T, deps api.ResolverDependencies, req *http.Request) FixedAccessTokenData {
		t.Helper()

		accountID := fixedSystemAdminEmailAddressAccountIDMap[emailAddress]
		if accountID == "" {
			t.Fatalf("unknown SystemAdministrator email address: %s", emailAddress)
		}

		accessTokenData := FixedAccessTokenData{
			AccountID: uuid.Must(uuid.FromString(accountID)),
			Role:      types.RoleSystemAdministrator,
		}

		if organisationIDStr, ok := fixedAccountOrganisationIDMap[accountID]; ok {
			organisationID := uuid.Must(uuid.FromString(organisationIDStr))
			accessTokenData.OrganisationID = &organisationID
		}

		return createAuthSessionAndAddToRequest(t, deps, req, accessTokenData)
	}
}

func ApplyAuthValuesFuncOrganisationAdministrator(emailAddress string) ApplyAuthValuesFunc {
	return func(t *testing.T, deps api.ResolverDependencies, req *http.Request) FixedAccessTokenData {
		t.Helper()

		accountID := fixedOrganisationAdminEmailAddressAccountIDMap[emailAddress]
		if accountID == "" {
			t.Fatalf("unknown OrganisationAdministrator email address: %s", emailAddress)
		}

		accessTokenData := FixedAccessTokenData{
			AccountID: uuid.Must(uuid.FromString(accountID)),
			Role:      types.RoleOrganisationAdministrator,
		}

		if organisationIDStr, ok := fixedAccountOrganisationIDMap[accountID]; ok {
			organisationID := uuid.Must(uuid.FromString(organisationIDStr))
			accessTokenData.OrganisationID = &organisationID
		}

		return createAuthSessionAndAddToRequest(t, deps, req, accessTokenData)
	}
}

func createAuthSessionAndAddToRequest(t *testing.T, deps api.ResolverDependencies, req *http.Request, accessTokenData FixedAccessTokenData) FixedAccessTokenData {
	t.Helper()

	expiresAt := authentication.TokenExpiryTypeDefault.GetExpiresAt(deps.TimeSource)
	cmd, err := command.NewAuthSessionCreateCmd(accessTokenData.AccountID, expiresAt)
	if err != nil {
		t.Fatalf("failed to create auth session create command: %v", err)
	}

	ctx := GetContextWithSystemAdministrator()
	err = deps.Handler().AuthSessionCreate(ctx, cmd)
	if err != nil {
		t.Fatalf("failed to create auth session: %v", err)
	}

	addAccessTokenToRequest(t, deps.TimeSource, req, accessTokenData, cmd.AuthAccessTokenID)
	addCsrfTokenToRequest(t, deps.TimeSource, req, cmd.AuthAccessTokenID)

	accessTokenData.AuthSessionID = cmd.AuthSessionID

	return accessTokenData
}

func GetContextWithSystemAdministrator() context.Context {
	ctx := context.Background()

	return authentication.WithAuthContext(ctx, authentication.AuthContext{
		Authenticated: true,
		Role:          types.RoleSystemAdministrator,
	})
}

func addAccessTokenToRequest(t *testing.T, timeSource types.TimeSource, req *http.Request, accessTokenData FixedAccessTokenData, accessTokenID uuid.UUID) {
	t.Helper()

	accessTokenOpts := authentication.TokenOptsForAccount(accessTokenID, authentication.TokenExpiryTypeDefault)
	accessToken, err := authentication.GenerateAccessToken(FixedConfig, accessTokenData, timeSource, accessTokenOpts)
	if err != nil {
		t.Fatalf("failed to generate auth token: %v", err)
	}

	accessTokenCookie := http.Cookie{
		Name:  "access_token",
		Value: accessToken,
	}
	req.AddCookie(&accessTokenCookie)
}

func addCsrfTokenToRequest(t *testing.T, timeSource types.TimeSource, req *http.Request, accessTokenID uuid.UUID) {
	t.Helper()

	csrfTokenOpts := authentication.TokenOptsForAccount(accessTokenID, authentication.TokenExpiryTypeDefault)
	csrfToken, err := authentication.GenerateCsrfToken(FixedConfig, timeSource, csrfTokenOpts)
	if err != nil {
		t.Fatalf("failed to generate csrf token: %v", err)
	}

	req.Header.Set("X-CSRF-Token", csrfToken)
}
