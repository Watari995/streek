package auth

import (
	"context"
	"database/sql"

	"github.com/Watari995/streek/backend/internal/apperror"
	domainAuth "github.com/Watari995/streek/backend/internal/domain/auth"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type Login struct {
	userRepo       repository.IUserRepository
	hasher         domainAuth.IPasswordHasher
	tokenGenerator domainAuth.ITokenGenerator
}

type LoginInput struct {
	Email    valueobject.Email
	Password valueobject.Password
}

type LoginOutput struct {
	AccessToken string
}

func NewLogin(userRepo repository.IUserRepository, hasher domainAuth.IPasswordHasher, tokenGenerator domainAuth.ITokenGenerator) *Login {
	return &Login{
		userRepo:       userRepo,
		hasher:         hasher,
		tokenGenerator: tokenGenerator,
	}
}

func (l *Login) Do(ctx context.Context, input LoginInput) (LoginOutput, error) {
	// find user by email
	user, err := l.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LoginOutput{}, errors.Wrap(err, "failed to find user by email")
	}
	// to prevent email enumeration, return the same response as a wrong password
	if user == nil {
		return LoginOutput{}, apperror.NewUnauthorizedError().SetMessage("invalid email or password")
	}
	// verify password
	err = l.hasher.Verify(input.Password, user.PasswordHash())
	if err != nil {
		return LoginOutput{}, apperror.NewUnauthorizedError().SetMessage("invalid email or password")
	}
	// generate access token
	accessToken, err := l.tokenGenerator.Generate(user.ID())
	if err != nil {
		return LoginOutput{}, errors.Wrap(err, "failed to generate access token")
	}

	return LoginOutput{
		AccessToken: accessToken,
	}, nil
}
