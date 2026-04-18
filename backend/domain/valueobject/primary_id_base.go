package valueobject

import (
	"database/sql/driver"
	"fmt"
	"strconv"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/gofrs/uuid/v5"
)

type PrimaryIDBase struct {
	uuid.UUID
	value string
}

func (e PrimaryIDBase) String() string {
	return e.value
}

func newPrimaryIDBase() PrimaryIDBase {
	u := uuid.Must(uuid.NewV7())
	return PrimaryIDBase{
		UUID:  u,
		value: u.String(),
	}
}

func newPrimaryIDBaseFromString(value string) (PrimaryIDBase, error) {
	if err := validation.Validate(value, validation.Required, is.UUID); err != nil {
		return PrimaryIDBase{}, err
	}
	uu, err := uuid.FromString(value)

	if err != nil {
		return PrimaryIDBase{}, errors.Wrap(err, "failed to parse uuid")
	}
	return PrimaryIDBase{UUID: uu, value: uu.String()}, nil
}

func (e *PrimaryIDBase) UnmarshalParam(src string) error {
	s := PrimaryIDBase{
		value: src,
	}

	err := validation.ValidateStruct(&s,
		validation.Field(&s.value, validation.Required, is.UUID),
	)

	if err != nil {
		return err
	}

	uu, err := uuid.FromString(src)
	e.value = uu.String()
	e.UUID = uu

	return err
}

func (e PrimaryIDBase) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

func (e *PrimaryIDBase) UnmarshalJSON(src []byte) error {
	str, err := strconv.Unquote(string(src))

	if err != nil {
		return err
	}

	s := PrimaryIDBase{
		value: str,
	}

	err = validation.ValidateStruct(&s, validation.Field(&s.value, validation.Required, is.UUID))

	if err != nil {
		return err
	}

	uu, err := uuid.FromString(str)
	e.UUID = uu
	e.value = uu.String()
	return err
}

func (e PrimaryIDBase) Value() (driver.Value, error) {
	return e.value, nil
}
