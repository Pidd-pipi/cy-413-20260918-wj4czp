package util

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
func WrapEntity(entity, field string, id uint, code int, err error) *AppError {
	return NewAppError(code, fmt.Sprintf("%s[id=%d] update failed: %s", entity, id, field), err)
}
