package point

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type GetBalance struct {
	pointLedgerRepo repository.IPointLedgerRepository
}

type GetBalanceInput struct {
	UserID valueobject.UserID
}

type GetBalanceOutput struct {
	Balance int
}

func NewGetBalance(pointLedgerRepo repository.IPointLedgerRepository) *GetBalance {
	return &GetBalance{pointLedgerRepo: pointLedgerRepo}
}

func (s *GetBalance) Do(ctx context.Context, input GetBalanceInput) (GetBalanceOutput, error) {
	balance, err := s.pointLedgerRepo.GetBalance(ctx, input.UserID)
	if err != nil {
		return GetBalanceOutput{}, errors.Wrap(err, "failed to get point balance")
	}
	return GetBalanceOutput{Balance: balance}, nil
}
