package controller

import "fmt"

// Error is an error that has a type
type Error struct {
	Type string
	Msg  string
}

// NewError returns a new error
func NewError(kind string, err error) *Error {
	return &Error{kind, err.Error()}
}

func (err Error) Error() string {
	return fmt.Sprintf("%s %s", err.Type, err.Msg)
}
