package helper

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt hash cost (defaults to 10), but 12 is more secure
const hashCost = 12

func GenerateHashFromPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, hashCost)
}

func CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

// DefaultHashForComparison returns a default password hash (empty string)
// with default hash cost for constant comparison of missing accounts
func DefaultHashForComparison() []byte {
	return []byte(fmt.Sprintf("$2a$%d$OYBKkpJa62bfLC011fuZNeSPZ3ensWQ7WwiHe/P1oP7bXwQ841pUa", hashCost))
}
