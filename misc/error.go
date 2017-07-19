package misc

import (
	"fmt"
)

var (
	NotImplementedErr         = &Error{Code: -1, Msg: "not implemented", IsPublic: false}
	InvalidArgumentsErr       = &Error{Code: 100, Msg: "invalid arguments", IsPublic: false}
	idGenerationErr           = &Error{Code: 101, Msg: "Failed to generate id", IsPublic: false}
	InsufficientPermissionErr = &Error{Code: 102, Msg: "insufficient permissions", IsPublic: true}
	InvalidOperationErr       = &Error{Code: 103, Msg: "invalid operation", IsPublic: true}
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
