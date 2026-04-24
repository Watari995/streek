package apperror

type MyError interface {
	Error() string // satisfies the built-in error interface
	Message() string
	Code() ErrorCode
	Status() int
	SetMessage(msg string) MyError
	SetCode(code ErrorCode) MyError
}
