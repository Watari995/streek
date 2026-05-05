package valueobject

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PositiveInt struct {
	value int
}

func NewPositiveInt(v int) (PositiveInt, error) {
	err := validation.Validate(v,
		validation.Min(1),
	)
	if err != nil {
		return PositiveInt{}, err
	}
	return PositiveInt{value: v}, nil
}

func (p PositiveInt) Int() int {
	return p.value
}

func MustPositiveInt(v int) PositiveInt {
	vo, err := NewPositiveInt(v)
	if err != nil {
		panic(err)
	}
	return vo
}
