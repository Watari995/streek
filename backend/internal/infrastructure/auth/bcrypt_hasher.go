package auth

import (
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(password valueobject.Password) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password.PlainText()), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptHasher) Verify(password valueobject.Password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password.PlainText()))
}
