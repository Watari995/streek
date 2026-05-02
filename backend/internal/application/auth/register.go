package auth

import (
	"context"
	"database/sql"

	"github.com/Watari995/streek/backend/internal/apperror"
	domainAuth "github.com/Watari995/streek/backend/internal/domain/auth"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type Register struct {
	userRepo       repository.IUserRepository
	hasher         domainAuth.IPasswordHasher
	tokenGenerator domainAuth.ITokenGenerator
}

type RegisterInput struct {
	Email    valueobject.Email
	Password valueobject.Password
}

type RegisterOutput struct {
	AccessToken string
}

func NewRegister(
	userRepo repository.IUserRepository,
	hasher domainAuth.IPasswordHasher,
	tokenGenerator domainAuth.ITokenGenerator,
) *Register {
	return &Register{
		userRepo:       userRepo,
		hasher:         hasher,
		tokenGenerator: tokenGenerator,
	}
}

func (r *Register) Do(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	// email conflict check
	existingUser, err := r.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RegisterOutput{}, errors.Wrap(err, "failed to find user by email")
	}
	if existingUser != nil {
		return RegisterOutput{}, apperror.NewConflictError().SetMessage("email already in use")
	}
	// hash password
	hashedPassword, err := r.hasher.Hash(input.Password)
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to hash password")
	}

	// create user
	user := entity.CreateUser(input.Email, hashedPassword)
	savedUser, err := r.userRepo.Save(ctx, user)
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to save user")
	}

	// generate access token
	accessToken, err := r.tokenGenerator.Generate(savedUser.ID())
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to generate access token")
	}

	return RegisterOutput{
		AccessToken: accessToken,
	}, nil
}
