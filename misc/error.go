package misc

import (
	"fmt"
)

var (
	NotImplementedErr   = &Error{Code: -1, Msg: "not implemented", IsPublic: false}
	NilCriticalParamErr = &Error{Code: 0, Msg: "nil critical param", IsPublic: false}
	idGenerationErr     = &Error{Code: 1, Msg: "Failed to generate id", IsPublic: false}
)

type PermissionedError interface {
	IsPrivate() bool
}

type Error struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	IsPublic bool   `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func (e *Error) IsPrivate() bool {
	return !e.IsPublic
}

type ErrorRef struct {
	Id Id `json:"id"`
}

func (e *ErrorRef) Error() string {
	return fmt.Sprintf("errorRef: %s", e.Id.String())
}
