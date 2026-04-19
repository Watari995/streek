package valueobject

import validation "github.com/go-ozzo/ozzo-validation/v4"

type String50 struct {
	value string
}

func NewString50(s string) (*String50, error) {
	err := validation.Validate(
		s,
		validation.Required,
		validation.RuneLength(1, 50),
	)
	if err != nil {
		return nil, err
	}

	return &String50{value: s}, nil
}

func (s String50) String() string {
	return s.value
}
