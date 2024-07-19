package auth

import (
	"context"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
)

//nolint:gochecknoglobals
var (
	fixedSystemAdminAccountID = uuid.Must(uuid.FromString("d7037ad0-d4bb-4dcc-8759-d82fbb3354e8"))

	fixedOrganisationAdminAccountID = uuid.Must(uuid.FromString("3ad082c7-cbda-49e1-a707-c53e1962be65"))
	fixedOrganisationID             = uuid.Must(uuid.FromString("6330de58-2761-411e-a243-bec6d0c53876"))

	fixedTokenSecret = "f71ab8929ad747915e135b8e9a5e01403329cc6b202c8e540e74920a78394e36f6266e4a505bf9cd362206bfd39665c69330e038f96ba72bbbc1f4a522564410" //nolint:gosec
)

type ApplyAuthValuesFunc func(t *testing.T, timeSource domain.TimeSource, req *http.Request) FixedAuthTokenData

func ApplyFixedAuthValuesOrganisationAdministrator(t *testing.T, timeSource domain.TimeSource, req *http.Request) FixedAuthTokenData {
	t.Helper()

	authTokenData := FixedAuthTokenData{
		TokenSecret:    mustHexDecode(fixedTokenSecret),
		AccountID:      fixedOrganisationAdminAccountID,
		OrganisationID: uuid.NullUUID{Valid: true, UUID: fixedOrganisationID},
		RoleIdentifier: string(domain.RoleOrganisationAdministrator),
	}

	addTokenToRequest(t, timeSource, req, authTokenData)

	return authTokenData
}

var _ ApplyAuthValuesFunc = ApplyFixedAuthValuesOrganisationAdministrator

func ApplyFixedAuthValuesSystemAdministrator(t *testing.T, timeSource domain.TimeSource, req *http.Request) FixedAuthTokenData {
	t.Helper()

	authTokenData := FixedAuthTokenData{
		TokenSecret:    mustHexDecode(fixedTokenSecret),
		AccountID:      fixedSystemAdminAccountID,
		RoleIdentifier: string(domain.RoleSystemAdministrator),
	}

	addTokenToRequest(t, timeSource, req, authTokenData)

	return authTokenData
}

var _ ApplyAuthValuesFunc = ApplyFixedAuthValuesSystemAdministrator

func addTokenToRequest(t *testing.T, timeSource domain.TimeSource, req *http.Request, authTokenData FixedAuthTokenData) {
	t.Helper()

	tokenOpts := authentication.TokenOptsForAccount(authTokenData, false)
	authToken, err := authentication.GenerateAuthToken(authTokenData, timeSource, tokenOpts)
	if err != nil {
		t.Fatalf("failed to generate auth token: %v", err)
	}

	csrfToken, err := authentication.GenerateCsrfToken(authTokenData, timeSource, tokenOpts)
	if err != nil {
		t.Fatalf("failed to generate CSRF token: %v", err)
	}

	req.Header.Set("X-CSRF-Token", csrfToken)
	cookie := http.Cookie{
		Name:  "authToken",
		Value: authToken,
	}
	req.AddCookie(&cookie)
}

func mustHexDecode(secret string) []byte {
	data, err := hex.DecodeString(secret)
	if err != nil {
		panic(err)
	}
	return data
}

func GetContextWithSystemAdministrator() context.Context {
	ctx := context.Background()

	return authentication.WithAuthContext(ctx, authentication.AuthContext{
		Authenticated: true,
		Role:          domain.RoleSystemAdministrator,
	})
}

func GetContextWithOrganisationAdministrator() context.Context {
	ctx := context.Background()

	organisationID := fixedOrganisationID
	return authentication.WithAuthContext(ctx, authentication.AuthContext{
		Authenticated:  true,
		OrganisationID: &organisationID,
		Role:           domain.RoleOrganisationAdministrator,
	})
}

func GetContextWithAnonymous() context.Context {
	ctx := context.Background()

	return authentication.WithAuthContext(ctx, authentication.AuthContext{
		Authenticated: false,
	})
}
