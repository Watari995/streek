package eventhandler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Watari995/streek/backend/internal/application/eventhandler"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/event/types"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPointLedgerRepository struct {
	saveFunc                  func(ctx context.Context, e entity.PointLedger) (*entity.PointLedger, error)
	getBalanceFunc            func(ctx context.Context, userID valueobject.UserID) (int, error)
	findByUserIDFunc          func(ctx context.Context, userID valueobject.UserID) ([]*entity.PointLedger, error)
	existsByIdempotencyKeyFn  func(ctx context.Context, key string) (bool, error)
	saveCalls                 []entity.PointLedger
}

func (m *mockPointLedgerRepository) Save(ctx context.Context, e entity.PointLedger) (*entity.PointLedger, error) {
	m.saveCalls = append(m.saveCalls, e)
	if m.saveFunc != nil {
		return m.saveFunc(ctx, e)
	}
	return &e, nil
}

func (m *mockPointLedgerRepository) GetBalance(ctx context.Context, userID valueobject.UserID) (int, error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(ctx, userID)
	}
	return 0, nil
}

func (m *mockPointLedgerRepository) FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*entity.PointLedger, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockPointLedgerRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	if m.existsByIdempotencyKeyFn != nil {
		return m.existsByIdempotencyKeyFn(ctx, key)
	}
	return false, nil
}

func newCheckInCompletedEvent(t *testing.T) types.CheckInCompletedEvent {
	t.Helper()
	checkedDate, err := valueobject.NewDateStringFromString("2026-05-05")
	require.NoError(t, err)
	return types.NewCheckInCompletedEvent(
		valueobject.NewUserID(),
		valueobject.NewHabitID(),
		checkedDate,
		lo.Must(valueobject.NewPositiveInt(10)),
		lo.Must(valueobject.NewString50("CHECK_IN")),
		"checkin:abc:2026-05-05",
		time.Now(),
	)
}

func TestEarnPointsOnCheckIn_Handle_SavesPointLedger(t *testing.T) {
	t.Parallel()
	repo := &mockPointLedgerRepository{}
	h := eventhandler.NewEarnPointsOnCheckIn(repo)
	e := newCheckInCompletedEvent(t)

	err := h.Handle(context.Background(), e)
	require.NoError(t, err)
	require.Len(t, repo.saveCalls, 1)

	saved := repo.saveCalls[0]
	assert.Equal(t, e.UserID.String(), saved.UserID().String())
	require.NotNil(t, saved.HabitID())
	assert.Equal(t, e.HabitID.String(), saved.HabitID().String())
	assert.Equal(t, e.PointAmount.Int(), saved.Amount().Int())
	assert.Equal(t, e.PointReason.String(), saved.Reason().String())
	assert.Equal(t, e.IdempotencyKey, saved.IdempotencyKey())
	assert.True(t, saved.PointType().IsEarn())
}

type unknownEvent struct{}

func (unknownEvent) EventType() string     { return "unknown" }
func (unknownEvent) OccurredAt() time.Time { return time.Time{} }

func TestEarnPointsOnCheckIn_Handle_RejectsUnexpectedEventType(t *testing.T) {
	t.Parallel()
	repo := &mockPointLedgerRepository{}
	h := eventhandler.NewEarnPointsOnCheckIn(repo)

	err := h.Handle(context.Background(), unknownEvent{})
	require.Error(t, err)
	assert.Empty(t, repo.saveCalls, "should not call Save on type mismatch")
}

func TestEarnPointsOnCheckIn_Handle_PropagatesRepoError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	repo := &mockPointLedgerRepository{
		saveFunc: func(ctx context.Context, e entity.PointLedger) (*entity.PointLedger, error) {
			return nil, repoErr
		},
	}
	h := eventhandler.NewEarnPointsOnCheckIn(repo)

	err := h.Handle(context.Background(), newCheckInCompletedEvent(t))
	require.Error(t, err)
}
