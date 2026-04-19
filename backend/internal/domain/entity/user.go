package entity

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type User struct {
	id           valueobject.UserID
	email        valueobject.Email
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

func (u *User) ID() valueobject.UserID {
	return u.id
}

func (u *User) Email() valueobject.Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// make user entity from db model
func NewUser(
	id valueobject.UserID,
	email valueobject.Email,
	passwordHash string,
	createdAt time.Time,
	updatedAt time.Time,
) User {
	return User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// make user entity from params
func CreateUser(
	email valueobject.Email,
	passwordHash string,
) User {
	return NewUser(
		valueobject.NewUserID(),
		email,
		passwordHash,
		time.Now(),
		time.Now(),
	)
}
