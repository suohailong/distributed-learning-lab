package error

import (
	"strconv"
	"strings"
)

type HError struct {
	Code    int
	Message string
}

func (e *HError) Error() string {
	var b strings.Builder
	code := strconv.Itoa(e.Code)
	b.WriteString("Error: ")
	b.WriteString(code)
	b.WriteString(" - ")
	b.WriteString(e.Message)
	s := b.String()
	return s

}
