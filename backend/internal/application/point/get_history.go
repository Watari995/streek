package point

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type GetHistory struct {
	pointLedgerRepo repository.IPointLedgerRepository
}

type GetHistoryInput struct {
	UserID valueobject.UserID
}

type GetHistoryOutput struct {
	Entries []*entity.PointLedger
}

func NewGetHistory(pointLedgerRepo repository.IPointLedgerRepository) *GetHistory {
	return &GetHistory{pointLedgerRepo: pointLedgerRepo}
}

func (s *GetHistory) Do(ctx context.Context, input GetHistoryInput) (GetHistoryOutput, error) {
	entries, err := s.pointLedgerRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return GetHistoryOutput{}, errors.Wrap(err, "failed to get point history")
	}
	return GetHistoryOutput{Entries: entries}, nil
}
