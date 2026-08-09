package protocol

import "fmt"

type ErrorCode string

const (
	ErrInvalidVersion  ErrorCode = "invalid_protocol_version"
	ErrInvalidType     ErrorCode = "invalid_message_type"
	ErrInvalidPayload  ErrorCode = "invalid_payload"
	ErrMalformedJSON   ErrorCode = "malformed_json"
	ErrValidation      ErrorCode = "validation_failed"
	ErrDispatch        ErrorCode = "dispatch_error"
	// ErrRelationshipRequired is returned when a message targets a user with
	// whom the sender has no accepted chat relationship. Clients should
	// surface this as "request pending," not a generic failure.
	ErrRelationshipRequired ErrorCode = "relationship_required"
)

// Error is a structured protocol error sent to clients.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, msg string) *Error {
	return &Error{Code: code, Message: msg}
}
