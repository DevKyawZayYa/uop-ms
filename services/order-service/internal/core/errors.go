package core

import "net/http"

type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

func New(code, msg string, status int) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: msg,
	}
}

func NewInternal(code, msg string) *AppError {
	return New(code, msg, http.StatusInternalServerError)
}

func NewBadRequest(code, msg string) *AppError {
	return New(code, msg, http.StatusBadRequest)
}

func NewConflict(code, msg string) *AppError {
	return New(code, msg, http.StatusConflict)
}

func NewServiceUnavailable(code, msg string) *AppError {
	return New(code, msg, http.StatusServiceUnavailable)
}
