package service

var (
	ErrUserNotFound  = newErr("user not found")
	ErrInvalidStatus = newErr("invalid user status")
)

type userErr struct {
	message string
}

func (u *userErr) Error() string {
	return u.message
}

func newErr(msg string) *userErr {
	return &userErr{message: msg}
}
