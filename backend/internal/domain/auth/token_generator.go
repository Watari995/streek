package auth

import "github.com/Watari995/streek/backend/internal/domain/valueobject"

type ITokenGenerator interface {
	Generate(userID valueobject.UserID) (string, error)
	Validate(token string) (valueobject.UserID, error)
}
