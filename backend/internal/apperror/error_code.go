package apperror

type ErrorCode string

const (
	CodeBadRequest    ErrorCode = "BAD_REQUEST"
	CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	CodeForbidden     ErrorCode = "FORBIDDEN"
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeConflict      ErrorCode = "CONFLICT"
	CodeInternalError ErrorCode = "INTERNAL_ERROR"
)
