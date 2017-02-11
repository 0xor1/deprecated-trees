package misc

import (
	"fmt"
)

var (
	NotImplementedErr = &Error{Code: -1, Msg: "not implemented"}
	idGenerationErr   = &Error{Code: 1, Msg: "Failed to generate id"}
)

func NilCriticalParamPanic(paramName string) {
	panic(&Error{Code: 0, Msg:  fmt.Sprintf("nil %s", paramName)})
}

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

type ErrorRef struct {
	Id Id `json:"id"`
}

func (e *ErrorRef) Error() string {
	return fmt.Sprintf("errorRef: %s", e.Id.String())
}
