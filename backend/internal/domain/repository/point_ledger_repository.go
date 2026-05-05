package repository

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type IPointLedgerRepository interface {
	Save(context.Context, entity.PointLedger) (*entity.PointLedger, error)
	GetBalance(context.Context, valueobject.UserID) (int, error)
	FindByUserID(context.Context, valueobject.UserID) ([]*entity.PointLedger, error)
	ExistsByIdempotencyKey(context.Context, string) (bool, error)
}
