package valueobject

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type DateString struct {
	value string
}

// YYYY-MM-DD format
var dateStringPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func NewDateString(s string) (*DateString, error) {
	err := validation.Validate(s,
		validation.Required,
		validation.Match(dateStringPattern),
	)
	if err != nil {
		return nil, err
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return nil, err
	}
	return &DateString{value: s}, nil
}

func (d DateString) String() string {
	return d.value
}
