package database_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/Watari995/streek/backend/internal/infrastructure/database"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPointLedgerEntity(t *testing.T, withHabit bool) entity.PointLedger {
	t.Helper()
	userID := valueobject.NewUserID()
	var habitID *valueobject.HabitID
	if withHabit {
		hid := valueobject.NewHabitID()
		habitID = &hid
	}
	return entity.CreatePointLedger(
		userID,
		habitID,
		valueobject.NewPointTypeEarn(),
		lo.Must(valueobject.NewPositiveInt(10)),
		lo.Must(valueobject.NewString50("CHECK_IN")),
		"checkin:abc:2026-05-05",
	)
}

func TestPointLedgerRepository_Save_InsertsRow(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	pl := newPointLedgerEntity(t, true)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO point_ledger")).
		WithArgs(
			pl.ID().String(),
			pl.UserID().String(),
			pl.HabitID().String(),
			pl.PointType().String(),
			pl.Amount().Int(),
			pl.Reason().String(),
			pl.IdempotencyKey(),
			pl.CreatedAt(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	got, err := repo.Save(context.Background(), pl)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_Save_HandlesNilHabitID(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	pl := newPointLedgerEntity(t, false) // habitID nil
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO point_ledger")).
		WithArgs(
			pl.ID().String(),
			pl.UserID().String(),
			nil, // habit_id is NULL
			pl.PointType().String(),
			pl.Amount().Int(),
			pl.Reason().String(),
			pl.IdempotencyKey(),
			pl.CreatedAt(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := repo.Save(context.Background(), pl)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_Save_PropagatesDBError(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	pl := newPointLedgerEntity(t, true)
	dbErr := errors.New("connection lost")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO point_ledger")).
		WillReturnError(dbErr)

	_, err := repo.Save(context.Background(), pl)
	require.ErrorIs(t, err, dbErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_GetBalance_ReturnsZeroForNoEntries(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	rows := sqlmock.NewRows([]string{"balance"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	got, err := repo.GetBalance(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 0, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_GetBalance_ReturnsEarnMinusSpend(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	// Simulating earn 30 - spend 12 = 18
	rows := sqlmock.NewRows([]string{"balance"}).AddRow(18)
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	got, err := repo.GetBalance(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 18, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_GetBalance_PropagatesDBError(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetBalance(context.Background(), userID)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_FindByUserID_ReturnsEntries(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	habitID := valueobject.NewHabitID()
	now := time.Now()
	entryID := valueobject.NewPointLedgerID()
	habitIDStr := habitID.String()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "habit_id", "type", "amount",
		"reason", "idempotency_key", "created_at",
	}).AddRow(
		entryID.String(), userID.String(), habitIDStr, "EARN", 10,
		"CHECK_IN", "checkin:abc:2026-05-05", now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	got, err := repo.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, userID.String(), got[0].UserID().String())
	require.NotNil(t, got[0].HabitID())
	assert.Equal(t, habitIDStr, got[0].HabitID().String())
	assert.Equal(t, 10, got[0].Amount().Int())
	assert.True(t, got[0].PointType().IsEarn())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_FindByUserID_HandlesNullHabitID(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	now := time.Now()
	entryID := valueobject.NewPointLedgerID()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "habit_id", "type", "amount",
		"reason", "idempotency_key", "created_at",
	}).AddRow(
		entryID.String(), userID.String(), nil, "SPEND", 5,
		"REDEEM", "redeem:xyz", now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	got, err := repo.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Nil(t, got[0].HabitID())
	assert.True(t, got[0].PointType().IsSpend())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_FindByUserID_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	userID := valueobject.NewUserID()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "habit_id", "type", "amount",
		"reason", "idempotency_key", "created_at",
	})
	mock.ExpectQuery(regexp.QuoteMeta("FROM point_ledger")).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	got, err := repo.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_ExistsByIdempotencyKey_True(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	key := "checkin:abc:2026-05-05"
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(key).
		WillReturnRows(rows)

	got, err := repo.ExistsByIdempotencyKey(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_ExistsByIdempotencyKey_False(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	key := "missing:key"
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(key).
		WillReturnRows(rows)

	got, err := repo.ExistsByIdempotencyKey(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPointLedgerRepository_ExistsByIdempotencyKey_PropagatesDBError(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	repo := database.NewPointLedgerRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs("any").
		WillReturnError(sql.ErrConnDone)

	_, err := repo.ExistsByIdempotencyKey(context.Background(), "any")
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}
