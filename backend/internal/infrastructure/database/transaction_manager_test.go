package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Watari995/streek/backend/internal/infrastructure/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "postgres"), mock
}

func TestTransactionManager_Run_CommitsOnSuccess(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	mock.ExpectBegin()
	mock.ExpectCommit()

	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_RollsBackOnError(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	wantErr := errors.New("fn failed")
	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		return wantErr
	})

	require.ErrorIs(t, err, wantErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_ReusesExistingTx(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	// ネストしても Begin / Commit は1回ずつのみ
	mock.ExpectBegin()
	mock.ExpectCommit()

	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		return txMgr.Run(ctx, func(ctx context.Context) error {
			return nil
		})
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_RetriesOnDeadlock(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	// 1回目: deadlock で rollback
	mock.ExpectBegin()
	mock.ExpectRollback()
	// 2回目: 成功で commit
	mock.ExpectBegin()
	mock.ExpectCommit()

	deadlockErr := &pq.Error{Code: "40P01"}
	callCount := 0
	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return deadlockErr
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_NoRetryOnOtherError(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	otherErr := errors.New("non-deadlock error")
	callCount := 0
	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		callCount++
		return otherErr
	})

	require.ErrorIs(t, err, otherErr)
	assert.Equal(t, 1, callCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_ReturnsLastErrAfterMaxRetries(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	for i := 0; i < 3; i++ {
		mock.ExpectBegin()
		mock.ExpectRollback()
	}

	deadlockErr := &pq.Error{Code: "40P01"}
	callCount := 0
	txMgr := database.NewTransactionManager(db)
	err := txMgr.Run(context.Background(), func(ctx context.Context) error {
		callCount++
		return deadlockErr
	})

	var pqErr *pq.Error
	require.ErrorAs(t, err, &pqErr)
	assert.Equal(t, pq.ErrorCode("40P01"), pqErr.Code)
	assert.Equal(t, 3, callCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Run_RollsBackAndRePanicsOnPanic(t *testing.T) {
	t.Parallel()
	db, mock := newTestDB(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	txMgr := database.NewTransactionManager(db)
	assert.PanicsWithValue(t, "boom", func() {
		_ = txMgr.Run(context.Background(), func(ctx context.Context) error {
			panic("boom")
		})
	})
	require.NoError(t, mock.ExpectationsWereMet())
}
