package auth

import (
	"errors"
	"time"

	"github.com/Watari995/streek/backend/domain/valueobject"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// TODO: later replace it.
var secretKey = []byte("your-secret-key")

func GenerateToken(userID valueobject.UserID) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(tokenString string) (valueobject.UserID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return valueobject.UserID{}, err
	}

	if !token.Valid {
		return valueobject.UserID{}, errors.New("invalid token")
	}

	return valueobject.NewUserIDFromString(claims.UserID)
}
