package database

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

// save, find, delete

type checkInRow struct {
	ID          string `db:"id"`
	HabitID     string `db:"habit_id"`
	CheckedDate string `db:"checked_date"`
	CreatedAt   string `db:"created_at"`
}

type CheckInRepository struct {
	db *sqlx.DB
}

func NewCheckInRepository(db *sqlx.DB) *CheckInRepository {
	return &CheckInRepository{db: db}
}

func (r *CheckInRepository) Save(ctx context.Context, checkIn entity.CheckIn) (*entity.CheckIn, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO check_ins (id, habit_id, check_in_date, created_at)
		VALUES ($1, $2, $3, $4)
		`, checkIn.ID(), checkIn.HabitID(), checkIn.CheckedDate().Format("2006-01-02"), checkIn.CreatedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

func (r *CheckInRepository) FindByID(ctx context.Context, id valueobject.CheckInID) (*entity.CheckIn, error) {
	var row checkInRow
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, habit_id, check_in_date, created_at
		FROM check_ins
		WHERE id = $1`, id).StructScan(&row)

	if err != nil {
		return nil, err
	}

	return r.toEntity(row)
}

func (r *CheckInRepository) Delete(ctx context.Context, id valueobject.CheckInID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM check_ins
		WHERE id = $1`, id)
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

	checkedDate, err := valueobject.NewCheckedDateFromString(row.CheckedDate)
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
