package auth

import (
	"errors"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTGenerator struct {
	secretKey []byte
}

func NewJWTGenerator(secretKey []byte) *JWTGenerator {
	return &JWTGenerator{secretKey: secretKey}
}

func (g *JWTGenerator) Generate(userID valueobject.UserID) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secretKey)
}

func (g *JWTGenerator) Validate(tokenString string) (valueobject.UserID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return g.secretKey, nil
	})

	if err != nil {
		return valueobject.UserID{}, err
	}

	if !token.Valid {
		return valueobject.UserID{}, errors.New("invalid token")
	}

	return valueobject.NewUserIDFromString(claims.UserID)
}
