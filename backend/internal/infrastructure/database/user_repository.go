package database

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

type userRow struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user entity.User) (*entity.User, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID(), user.Email().String(), user.PasswordHash(),
		user.CreatedAt(), user.UpdatedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id valueobject.UserID) (*entity.User, error) {
	var row userRow
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).StructScan(&row)
	if err != nil {
		return nil, err
	}

	return r.toEntity(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error) {
	var row userRow
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		email.String(),
	).StructScan(&row)
	if err != nil {
		return nil, err
	}

	return r.toEntity(row)
}

func (r *UserRepository) toEntity(row userRow) (*entity.User, error) {
	userID, err := valueobject.NewUserIDFromString(row.ID)
	if err != nil {
		return nil, err
	}

	email, err := valueobject.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(userID, *email, row.PasswordHash, row.CreatedAt, row.UpdatedAt)
	return &user, nil
}
