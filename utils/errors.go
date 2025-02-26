package utils

import (
	"fmt"
	"log"
	"runtime"
)

// Error represents a custom error with additional context
type Error struct {
	Message string
	Cause   error
	File    string
	Line    int
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v (at %s:%d)", e.Message, e.Cause, e.File, e.Line)
	}
	return fmt.Sprintf("%s (at %s:%d)", e.Message, e.File, e.Line)
}

// NewError creates a new error with the current file and line number
func NewError(message string, cause error) *Error {
	_, file, line, _ := runtime.Caller(1)
	return &Error{
		Message: message,
		Cause:   cause,
		File:    file,
		Line:    line,
	}
}

// LogError logs an error with additional context
func LogError(message string, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("ERROR: %s: %v (at %s:%d)", message, err, file, line)
	}
}

// HandleErrorWithMessage logs an error and returns a user-friendly message
func HandleErrorWithMessage(err error, userMessage string) string {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("ERROR: %v (at %s:%d)", err, file, line)
		return userMessage
	}
	return ""
}