package misc

import (
	"database/sql"
	"fmt"
)

var (
	NotImplementedErr         = &Error{Code: "g_ni", Msg: "not implemented", Public: true}
	InvalidArgumentsErr       = &Error{Code: "g_ia", Msg: "invalid arguments", Public: true}
	idGenerationErr           = &Error{Code: "g_ig", Msg: "failed to generate id", Public: false}
	InsufficientPermissionErr = &Error{Code: "g_ip", Msg: "insufficient permissions", Public: true}
	InvalidOperationErr       = &Error{Code: "g_io", Msg: "invalid operation", Public: true}
	MaxEntityCountExceededErr = &Error{Code: "g_mece", Msg: "max entity count exceeded", Public: true}
)

type PermissionedError interface {
	IsPublic() bool
}

type Error struct {
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Public bool   `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %s, msg: %s", e.Code, e.Msg)
}

func (e *Error) IsPublic() bool {
	return e.Public
}

type ErrorRef struct {
	Id Id `json:"id"`
}

func (e *ErrorRef) Error() string {
	return fmt.Sprintf("errorRef: %s", e.Id.String())
}

func PanicIf(e error) {
	if e != nil {
		panic(e)
	}
}

func IsSqlErrNoRowsAndPanicIf(e error) bool {
	if e != nil {
		if e == sql.ErrNoRows {
			return true
		}
		panic(e)
	}
	return false
}
