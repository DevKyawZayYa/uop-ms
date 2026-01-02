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

func NewInternal(code, msg string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: code, Message: msg}
}
