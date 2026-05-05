package valueobject

import validation "github.com/go-ozzo/ozzo-validation/v4"

type PointType struct {
	value string
}

const (
	pointTypeEarn  = "EARN"
	pointTypeSpend = "SPEND"
)

func NewPointTypeEarn() PointType {
	return PointType{value: pointTypeEarn}
}

func NewPointTypeSpend() PointType {
	return PointType{value: pointTypeSpend}
}

func NewPointType(v string) (PointType, error) {
	err := validation.Validate(v,
		validation.Required,
		validation.In(pointTypeEarn, pointTypeSpend),
	)
	if err != nil {
		return PointType{}, err
	}
	return PointType{value: v}, nil
}

func (p PointType) String() string {
	return p.value
}

func (p PointType) IsEarn() bool {
	return p.value == pointTypeEarn
}

func (p PointType) IsSpend() bool {
	return p.value == pointTypeSpend
}
