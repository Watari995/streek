package value_object

import (
	"database/sql/driver"
	"fmt"
	"strconv"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid/v5"
)

type PrimaryIdBase struct {
	uuid.UUID
	value string
}

func (e PrimaryIdBase) String() string {
	return e.value
}

func newPrimaryIdBase() PrimaryIdBase {
	u := uuid.Must(uuid.NewV7())
	return PrimaryIdBase{
		UUID:  u,
		value: u.String(),
	}
}

func newPrimaryIdBaseFromString(value string) (PrimaryIdBase, error) {
	if err := validation.Validate(value, validation.Required, is.UUID); err != nil {
		return PrimaryIdBase{}, err
	}
	uu, err := uuid.FromString(value)

	if err != nil {
		return PrimaryIdBase{}, errors.Wrap(err, "failed to parse uuid")
	}
	return PrimaryIdBase{UUID: uu, value: uu.String()}, nil
}

func (e *PrimaryIdBase) UnmarshalParam(src string) error {
	s := PrimaryIdBase{
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

func (e PrimaryIdBase) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

func (e *PrimaryIdBase) UnmarshalJSON(src []byte) error {
	str, err := strconv.Unquote(string(src))

	if err != nil {
		return err
	}

	s := PrimaryIdBase{
		value: str,
	}

	err = validation.ValidateStruct(&s,
		validation.Field(&s.value, validation.Required, is.UUID),
	)

	if err != nil { return err
	}

	uu, err := uuid.FromString(str)
	e.UUID = uu
	e.value = uu.String()
	return err
}

func (e PrimaryIdBase) Value() (driver.Value, error) {
	return e.value, nil
}

func (e *PrimaryIdBase) Scan(src interface{}) error {
	switch src := src.(type) {
	case uuid.UUID:
		e.UUID = src
		e.value = src.String()
		return nil

	case []byte:
		u := uuid.UUID{}
		if len(src) == uuid.Size {
			if err := u.UnmarshalBinary(src); err != nil {
				return err
			}
		}
		if err := u.UnmarshalText(src); err != nil {
			return err
		}
		e.UUID = u
		e.value = u.String()
		return nil
	case string:
		uu, err := uuid.FromString(src)
		e.UUID = uu
		e.value = uu.String()
		return err
	}
	return nil
}

func (e PrimaryIdBase) Validate() error {
	return validation.Validate(e.value, validation.Required, is.UUID)
}
