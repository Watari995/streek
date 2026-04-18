package valueobject

import (
validation "github.com/go-ozzo/ozzo-validation/v4"
"kalonade.com/internal/my_ozzo"
)

type Email struct {
LiteralBase[string]
}

var EmailValidationRule = []validation.Rule{
validation.Required,
validation.RuneLength(1, 319),
my_ozzo.IsLooseEmailFormat,
}

func (e Email) Validate() error {
return validation.Validate(e.v, EmailValidationRule...)
}

func NewEmail(email string) (\*Email, error) {
e := Email{
LiteralBase: LiteralBase[string]{
v: email,
},
}
if err := e.Validate(); err != nil {
return nil, err
}

    return &e, nil

}
