package valueobject

import validation "github.com/go-ozzo/ozzo-validation/v4"

type String200 struct {
	value string
}

func NewString200(s string) (*String200, error) {
	err := validation.Validate(
		s,
		validation.Required,
		validation.RuneLength(1, 200),
	)
	if err != nil {
		return nil, err
	}

	return &String200{value: s}, nil
}

func (s String200) String() string {
	return s.value
}
