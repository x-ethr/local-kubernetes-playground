package users

import (
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// Hash will hash a given password. Uses [bcrypt.DefaultCost].
//
//   - For specific error checking, see the [golang.org/x/crypto/bcrypt] package.
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func Verify(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false
	} else if err != nil {
		slog.Error("Unexpected Error While Comparing Hash & Password", slog.String("error", err.Error()))
	}

	return err == nil
}
