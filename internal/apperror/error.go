package apperror

// Error is the base typed HTTP error. Controllers throw a subclass (or wrap
// one) and let httpx write the status + {"error": message} body.
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func New(statusCode int, message string) *Error {
	return &Error{StatusCode: statusCode, Message: message}
}
