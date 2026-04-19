package repository

import "github.com/Watari995/streek/backend/domain/entity"

type ICheckInRepository interface {
	save(entity.CheckIn)
}
