package herror

import (
	"strconv"
	"strings"
)

const (
	OK           = 1
	ErrPutFailed = -1
)

type hError struct {
	Code    int
	Message string
}

func New(code int, message string) error {
	return &hError{
		Code:    code,
		Message: message,
	}
}

func (e *hError) Error() string {
	var b strings.Builder
	code := strconv.Itoa(e.Code)
	b.WriteString("Error: ")
	b.WriteString(code)
	b.WriteString(" - ")
	b.WriteString(e.Message)
	s := b.String()
	return s

}
