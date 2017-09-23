package misc

import (
	"fmt"
)

var (
	NotImplementedErr         = &Error{Code: "g_ni", Msg: "not implemented", IsPublic: false}
	InvalidArgumentsErr       = &Error{Code: "g_ia", Msg: "invalid arguments", IsPublic: false}
	idGenerationErr           = &Error{Code: "g_ig", Msg: "Failed to generate id", IsPublic: false}
	InsufficientPermissionErr = &Error{Code: "g_ip", Msg: "insufficient permissions", IsPublic: true}
	InvalidOperationErr       = &Error{Code: "g_io", Msg: "invalid operation", IsPublic: true}
	MaxEntityCountExceededErr = &Error{Code: "g_mece", Msg: "max entity count exceeded", IsPublic: true}
)

type PermissionedError interface {
	IsPrivate() bool
}

type Error struct {
	Code     string `json:"code"`
	Msg      string `json:"msg"`
	IsPublic bool   `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %s, msg: %s", e.Code, e.Msg)
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
