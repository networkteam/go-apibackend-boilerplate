package helper

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHashFromPassword(password []byte, hashCost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, hashCost)
}

func CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

// DefaultHashForComparison returns a default password hash (empty string)
// with default hash cost for constant comparison of missing accounts
func DefaultHashForComparison(hashCost int) []byte {
	return []byte(fmt.Sprintf("$2a$%d$OYBKkpJa62bfLC011fuZNeSPZ3ensWQ7WwiHe/P1oP7bXwQ841pUa", hashCost))
}
