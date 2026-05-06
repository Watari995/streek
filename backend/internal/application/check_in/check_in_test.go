package checkin_test

import (
	"context"
	"errors"
	"testing"
	"time"

	checkin "github.com/Watari995/streek/backend/internal/application/check_in"
	domainCache "github.com/Watari995/streek/backend/internal/domain/cache"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/event"
	"github.com/Watari995/streek/backend/internal/domain/event/types"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockCheckInRepository struct {
	saveFunc  func(ctx context.Context, c entity.CheckIn) (*entity.CheckIn, error)
	saveCalls []entity.CheckIn
}

func (m *mockCheckInRepository) Save(ctx context.Context, c entity.CheckIn) (*entity.CheckIn, error) {
	m.saveCalls = append(m.saveCalls, c)
	if m.saveFunc != nil {
		return m.saveFunc(ctx, c)
	}
	return &c, nil
}
func (m *mockCheckInRepository) FindByHabitID(ctx context.Context, habitID valueobject.HabitID) ([]*entity.CheckIn, error) {
	return nil, nil
}
func (m *mockCheckInRepository) DeleteByHabitIDAndCheckedDate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error {
	return nil
}

type mockHabitRepository struct {
	findByIDFunc func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error)
}

func (m *mockHabitRepository) Save(ctx context.Context, h entity.Habit) (*entity.Habit, error) {
	return &h, nil
}
func (m *mockHabitRepository) FindByID(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockHabitRepository) FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*entity.Habit, error) {
	return nil, nil
}
func (m *mockHabitRepository) Delete(ctx context.Context, id valueobject.HabitID) error {
	return nil
}

type mockStreakCache struct {
	invalidateCalls int
}

func (m *mockStreakCache) Get(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) (*domainCache.StreakSnapshot, bool, error) {
	return nil, false, nil
}
func (m *mockStreakCache) Set(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString, snapshot domainCache.StreakSnapshot) error {
	return nil
}
func (m *mockStreakCache) Invalidate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error {
	m.invalidateCalls++
	return nil
}

type mockEventPublisher struct {
	publishFunc  func(ctx context.Context, e event.DomainEvent) error
	publishCalls []event.DomainEvent
}

func (m *mockEventPublisher) Subscribe(eventType string, handler func(context.Context, event.DomainEvent) error) error {
	return nil
}
func (m *mockEventPublisher) SubscribeAsync(eventType string, handler func(context.Context, event.DomainEvent) error) error {
	return nil
}
func (m *mockEventPublisher) Publish(ctx context.Context, e event.DomainEvent) error {
	m.publishCalls = append(m.publishCalls, e)
	if m.publishFunc != nil {
		return m.publishFunc(ctx, e)
	}
	return nil
}

// fakeTxManager runs fn directly and returns its result. Suitable for unit tests
// that do not need real tx behaviour.
type fakeTxManager struct {
	runCalls int
}

func (f *fakeTxManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	f.runCalls++
	return fn(ctx)
}

// --- helpers ---

func newHabit(t *testing.T, ownerID valueobject.UserID) *entity.Habit {
	t.Helper()
	color, err := valueobject.NewHexColor("#FF0000")
	require.NoError(t, err)
	h := entity.CreateHabit(
		ownerID,
		lo.Must(valueobject.NewString50("Workout")),
		nil,
		color,
	)
	return &h
}

func newInput(habitID valueobject.HabitID, userID valueobject.UserID) checkin.CheckInInput {
	checkedDate := lo.Must(valueobject.NewDateStringFromString("2026-05-05"))
	return checkin.CheckInInput{
		HabitID:     habitID,
		UserID:      userID,
		CheckedDate: checkedDate,
	}
}

// --- tests ---

func TestCheckIn_Do_SavesCheckInAndPublishesEvent(t *testing.T) {
	t.Parallel()
	userID := valueobject.NewUserID()
	habit := newHabit(t, userID)

	checkInRepo := &mockCheckInRepository{}
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return habit, nil
		},
	}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)

	err := svc.Do(context.Background(), newInput(habit.ID(), userID))
	require.NoError(t, err)

	assert.Equal(t, 1, txMgr.runCalls, "transaction Run should be called")
	require.Len(t, checkInRepo.saveCalls, 1)
	require.Len(t, publisher.publishCalls, 1)
	assert.Equal(t, 1, cache.invalidateCalls, "cache should be invalidated after commit")

	// verify event payload
	e, ok := publisher.publishCalls[0].(types.CheckInCompletedEvent)
	require.True(t, ok, "published event must be CheckInCompletedEvent")
	assert.Equal(t, userID.String(), e.UserID.String())
	assert.Equal(t, habit.ID().String(), e.HabitID.String())
	assert.Equal(t, "2026-05-05", e.CheckedDate.String())
	assert.Greater(t, e.PointAmount.Int(), 0)
}

func TestCheckIn_Do_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	t.Parallel()
	userID := valueobject.NewUserID()

	checkInRepo := &mockCheckInRepository{}
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return nil, nil // not found
		},
	}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)
	err := svc.Do(context.Background(), newInput(valueobject.NewHabitID(), userID))
	require.Error(t, err)

	assert.Equal(t, 0, txMgr.runCalls, "tx must not start when habit missing")
	assert.Empty(t, checkInRepo.saveCalls)
	assert.Empty(t, publisher.publishCalls)
	assert.Equal(t, 0, cache.invalidateCalls)
}

func TestCheckIn_Do_ReturnsForbiddenWhenHabitOwnedByDifferentUser(t *testing.T) {
	t.Parallel()
	requesterID := valueobject.NewUserID()
	ownerID := valueobject.NewUserID() // different owner
	habit := newHabit(t, ownerID)

	checkInRepo := &mockCheckInRepository{}
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return habit, nil
		},
	}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)
	err := svc.Do(context.Background(), newInput(habit.ID(), requesterID))
	require.Error(t, err)

	assert.Empty(t, checkInRepo.saveCalls)
	assert.Empty(t, publisher.publishCalls)
	assert.Equal(t, 0, txMgr.runCalls)
}

func TestCheckIn_Do_PropagatesHabitLookupError(t *testing.T) {
	t.Parallel()
	userID := valueobject.NewUserID()

	dbErr := errors.New("db down")
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return nil, dbErr
		},
	}
	checkInRepo := &mockCheckInRepository{}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)
	err := svc.Do(context.Background(), newInput(valueobject.NewHabitID(), userID))
	require.Error(t, err)

	assert.Empty(t, checkInRepo.saveCalls)
	assert.Empty(t, publisher.publishCalls)
}

func TestCheckIn_Do_DoesNotInvalidateCacheWhenSaveFails(t *testing.T) {
	t.Parallel()
	userID := valueobject.NewUserID()
	habit := newHabit(t, userID)

	saveErr := errors.New("save failed")
	checkInRepo := &mockCheckInRepository{
		saveFunc: func(ctx context.Context, c entity.CheckIn) (*entity.CheckIn, error) {
			return nil, saveErr
		},
	}
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return habit, nil
		},
	}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)
	err := svc.Do(context.Background(), newInput(habit.ID(), userID))
	require.Error(t, err)

	assert.Equal(t, 1, txMgr.runCalls)
	assert.Empty(t, publisher.publishCalls, "event must not be published when checkin save fails")
	assert.Equal(t, 0, cache.invalidateCalls, "cache must not be invalidated on tx failure")
}

func TestCheckIn_Do_PropagatesPublisherError(t *testing.T) {
	t.Parallel()
	userID := valueobject.NewUserID()
	habit := newHabit(t, userID)

	publishErr := errors.New("publish failed")
	checkInRepo := &mockCheckInRepository{}
	habitRepo := &mockHabitRepository{
		findByIDFunc: func(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
			return habit, nil
		},
	}
	cache := &mockStreakCache{}
	publisher := &mockEventPublisher{
		publishFunc: func(ctx context.Context, e event.DomainEvent) error {
			return publishErr
		},
	}
	txMgr := &fakeTxManager{}

	svc := checkin.NewCheckIn(checkInRepo, habitRepo, cache, publisher, txMgr)
	err := svc.Do(context.Background(), newInput(habit.ID(), userID))
	require.Error(t, err)

	assert.Equal(t, 1, txMgr.runCalls)
	require.Len(t, checkInRepo.saveCalls, 1, "checkin save was attempted before publish")
	assert.Equal(t, 0, cache.invalidateCalls, "cache must not be invalidated when publish fails")
}

// suppress unused linter warnings for time import in case something gets removed
var _ = time.Time{}
