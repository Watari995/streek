package auth

import "github.com/Watari995/streek/backend/internal/domain/valueobject"

type IPasswordHasher interface {
	Hash(password valueobject.Password) (string, error)
	Verify(password valueobject.Password, hash string) error
}
