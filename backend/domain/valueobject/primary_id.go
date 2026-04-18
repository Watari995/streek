package valueobject

type UserID struct {
	PrimaryIDBase
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

type HabitID struct {
	PrimaryIDBase
}

func NewHabitID() HabitID {
	return HabitID{newPrimaryIdBase()}
}

func NewHabitIDFromString(s string) (HabitID, error) {
	base, err := newPrimaryIdBaseFromString(s)
	if err != nil {
		return HabitID{}, err
	}
	return HabitID{base}, nil
}

type CheckInID struct {
	PrimaryIDBase
}

func NewCheckInID() CheckInID {
	return CheckInID{newPrimaryIdBase()}
}

func NewCheckInIDFromString(s string) (CheckInID, error) {
	base, err := newPrimaryIdBaseFromString(s)
	if err != nil {
		return CheckInID{}, err
	}
	return CheckInID{base}, nil
}
