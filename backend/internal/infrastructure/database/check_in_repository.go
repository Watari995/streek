package database

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

type checkInRow struct {
	ID          string    `db:"id"`
	HabitID     string    `db:"habit_id"`
	CheckedDate time.Time `db:"checked_date"`
	CreatedAt   time.Time `db:"created_at"`
}

type CheckInRepository struct {
	db *sqlx.DB
}

func NewCheckInRepository(db *sqlx.DB) *CheckInRepository {
	return &CheckInRepository{db: db}
}

func (r *CheckInRepository) Save(ctx context.Context, checkIn entity.CheckIn) (*entity.CheckIn, error) {
	conn := getConn(ctx, r.db)
	_, err := conn.ExecContext(ctx, `
		INSERT INTO check_ins (id, habit_id, checked_date, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (habit_id, checked_date) DO NOTHING
		`,
		checkIn.ID(), checkIn.HabitID(), checkIn.CheckedDate().String(), checkIn.CreatedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

func (r *CheckInRepository) FindByHabitID(ctx context.Context, habitID valueobject.HabitID) ([]*entity.CheckIn, error) {
	rows := []checkInRow{}
	conn := getConn(ctx, r.db)
	err := sqlx.SelectContext(ctx, conn, &rows, `
		SELECT id, habit_id, checked_date, created_at
		FROM check_ins
		WHERE habit_id = $1`, habitID)
	if err != nil {
		return nil, err
	}

	checkIns := make([]*entity.CheckIn, len(rows))
	for i, row := range rows {
		checkIn, err := r.toEntity(row)
		if err != nil {
			return nil, err
		}
		checkIns[i] = checkIn
	}

	return checkIns, nil
}

func (r *CheckInRepository) DeleteByHabitIDAndCheckedDate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error {
	conn := getConn(ctx, r.db)
	_, err := conn.ExecContext(ctx, `
		DELETE FROM check_ins
		WHERE habit_id = $1 AND checked_date = $2`,
		habitID, date.String(),
	)
	return err
}

func (r *CheckInRepository) toEntity(row checkInRow) (*entity.CheckIn, error) {
	checkInID, err := valueobject.NewCheckInIDFromString(row.ID)
	if err != nil {
		return nil, err
	}

	habitID, err := valueobject.NewHabitIDFromString(row.HabitID)
	if err != nil {
		return nil, err
	}

	checkedDate, err := valueobject.NewDateStringFromString(row.CheckedDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	checkIn := entity.NewCheckIn(
		checkInID,
		habitID,
		checkedDate,
		row.CreatedAt,
	)
	return &checkIn, nil
}
