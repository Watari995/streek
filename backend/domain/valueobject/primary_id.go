package valueobject

type UserID struct {
	PrimaryIdBase
}

func NewUserID() UserID {
	return UserID{newPrimaryIdBase()}
}

func NewUserIDFromString(s string) (UserID, error) {
	base, err := newPrimaryIdBaseFromString(s)
	if err != nil {
		return UserID{}, err
	}
	return UserID{base}, nil
}
