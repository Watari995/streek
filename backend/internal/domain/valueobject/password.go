package valueobject

import validation "github.com/go-ozzo/ozzo-validation/v4"

type Password struct {
	value string
}

func NewPassword(s string) (Password, error) {
	err := validation.Validate(s,
		validation.Required,
		validation.RuneLength(8, 72),
	)
	if err != nil {
		return Password{}, err
	}
	return Password{value: s}, nil
}

func (p Password) String() string {
	return "[REDACTED]"
}

func (p Password) PlainText() string {
	return p.value
}
