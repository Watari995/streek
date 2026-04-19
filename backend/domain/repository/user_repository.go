package repository

import (
	"context"

	"github.com/Watari995/streek/backend/domain/entity"
	"github.com/Watari995/streek/backend/domain/valueobject"
)

type IUserRepository interface {
	Save(context.Context, entity.User) (*entity.User, error)
	FindByID(context.Context, valueobject.UserID) (*entity.User, error)
	FindByEmail(context.Context, valueobject.Email) (*entity.User, error)
}
