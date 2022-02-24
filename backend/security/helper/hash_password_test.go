package helper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"myvendor.mytld/myproject/backend/security/helper"
)

func TestGenerateHashFromPassword(t *testing.T) {
	passwordHash, err := helper.GenerateHashFromPassword([]byte("myRandomPassword"), bcrypt.MinCost)
	require.NoError(t, err)

	err = helper.CompareHashAndPassword(passwordHash, []byte("myRandomPassword"))
	require.NoError(t, err)
}
