package valueobject

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type HexColor struct {
	value string
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func NewHexColor(s string) (*HexColor, error) {
	err := validation.Validate(s,
		validation.Required,
		validation.Match(hexColorPattern),
	)
	if err != nil {
		return nil, err
	}
	return &HexColor{value: s}, nil
}

func (h HexColor) String() string {
	return h.value
}
