package main

import "fmt"

type ErrorKind string

const (
	ErrorKindUser   ErrorKind = "user"
	ErrorKindSystem ErrorKind = "system"
)

type AppError struct {
	Kind    ErrorKind
	Stage   string
	Message string
	Hint    string
	Cause   error
}

func (e *AppError) Error() string {
	base := fmt.Sprintf("[%s][%s] %s", e.Kind, e.Stage, e.Message)
	if e.Hint != "" {
		base += " | hint: " + e.Hint
	}
	if e.Cause != nil {
		base += fmt.Sprintf(" | cause: %v", e.Cause)
	}
	return base
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewUserError(stage, message, hint string) error {
	return &AppError{
		Kind:    ErrorKindUser,
		Stage:   stage,
		Message: message,
		Hint:    hint,
	}
}

func NewSystemError(stage string, cause error, hint string) error {
	return &AppError{
		Kind:    ErrorKindSystem,
		Stage:   stage,
		Message: "operation failed",
		Hint:    hint,
		Cause:   cause,
	}
}
