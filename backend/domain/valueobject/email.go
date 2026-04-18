package valueobject

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Email struct {
	value string
}

func NewEmail(s string) (*Email, error) {
	err := validation.Validate(
		s,
		validation.Required,
		is.EmailFormat,
		validation.RuneLength(1, 255),
	)
	if err != nil {
		return nil, err
	}
	return &Email{value: s}, nil
}

func (e Email) String() string { return e.value }
