package database

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

type habitRow struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	LabelColor  string    `db:"label_color"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type HabitRepository struct {
	db *sqlx.DB
}

func NewHabitRepository(db *sqlx.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) Save(ctx context.Context, habit entity.Habit) (*entity.Habit, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO habits (id, user_id, name, description, label_color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		habit.ID(), habit.UserID(), habit.Name(), habit.Description(), habit.LabelColor(),
		habit.CreatedAt(), habit.UpdatedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &habit, nil
}

func (r *HabitRepository) FindByID(ctx context.Context, id valueobject.HabitID) (*entity.Habit, error) {
	var row habitRow
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, user_id, name, description, label_color, created_at, updated_at
		FROM habits
		WHERE id = $1`, id).StructScan(&row)

	if err != nil {
		return nil, err
	}

	return r.toEntity(row)
}

func (r *HabitRepository) FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*entity.Habit, error) {
	rows := []habitRow{}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, name, description, label_color, created_at, updated_at
		FROM habits
		WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}

	habits := make([]*entity.Habit, len(rows))
	for i, row := range rows {
		habit, err := r.toEntity(row)
		if err != nil {
			return nil, err
		}
		habits[i] = habit
	}

	return habits, nil
}

func (r *HabitRepository) Delete(ctx context.Context, id valueobject.HabitID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM habits WHERE id = $1`, id)
	return err
}

func (r *HabitRepository) toEntity(row habitRow) (*entity.Habit, error) {
	habitID, err := valueobject.NewHabitIDFromString(row.ID)
	if err != nil {
		return nil, err
	}

	userID, err := valueobject.NewUserIDFromString(row.UserID)
	if err != nil {
		return nil, err
	}

	name, err := valueobject.NewString50(row.Name)
	if err != nil {
		return nil, err
	}

	var description *valueobject.String200
	if row.Description != nil {
		desc, err := valueobject.NewString200(*row.Description)
		if err != nil {
			return nil, err
		}
		description =  desc
	}

	labelColor, err := valueobject.NewHexColor(row.LabelColor)
	if err != nil {
		return nil, err
	}

	habit := entity.NewHabit(
		habitID,
		userID,
		name,
		description,
		*labelColor,
		row.CreatedAt,
		row.UpdatedAt,
	)

	return &habit, nil
}
