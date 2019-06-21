package lib

import (
	"runtime"
)

type errors struct {
	Message	string
	File	string
	Line	int
}

func NewError(message string) error {
	_, file, line, _ := runtime.Caller(2)
	return &errors{
		Message:	message,
		File:	file,
		Line:	line,
	}
}

func (e *errors) Error() string {
	return  e.Message
}