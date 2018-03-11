package util

import (
	"database/sql"
	"fmt"
)

var (
	NoSuchRegionErr           = &AppError{Code: "g_nsr", Message: "no such region", Public: true}
	NotImplementedErr         = &AppError{Code: "g_ni", Message: "not implemented", Public: true}
	InvalidArgumentsErr       = &AppError{Code: "g_ia", Message: "invalid arguments", Public: true}
	InsufficientPermissionErr = &AppError{Code: "g_ip", Message: "insufficient permissions", Public: true}
	InvalidOperationErr       = &AppError{Code: "g_io", Message: "invalid operation", Public: true}
	InvalidEntityCountErr     = &AppError{Code: "g_iec", Message: "invalid entity count", Public: true}
	NoSuchEntityErr           = &AppError{Code: "g_nse", Message: "no such entity", Public: true}
	idParseErr                = &AppError{Code: "g_idp", Message: "failed to parse id", Public: true}
	invalidEndpointErr        = &AppError{Code: "g_ie", Message: "invalid endpoint", Public: false}
	unauthorizedErr           = &AppError{Code: "g_ns", Message: "unauthorized", Public: true}
	internalServerErr         = &AppError{Code: "g_is", Message: "internal server error", Public: true}
	externalAppErr            = &AppError{Code: "g_ea", Message: "external error occurred", Public: false}
)

type PermissionedError interface {
	error
	IsPublic() bool
	Panic()
}

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Public  bool   `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %q message: %q", e.Code, e.Message)
}

func (e *AppError) IsPublic() bool {
	return e.Public
}

func (e *AppError) Panic() {
	panic(e)
}

type externalError struct {
	AppError
	OriginalError error `json:"-"`
}

func (e *externalError) Error() string {
	return fmt.Sprintf("code: %q message: %q original error: %q", e.Code, e.Message, e.OriginalError.Error())
}

func (e *externalError) Panic() {
	panic(e)
}

type missingDlmErr struct {
	dlmKey  string
	reqPath string
}

func (e *missingDlmErr) Error() string {
	return fmt.Sprintf("missing dlm key %q on path %s", e.dlmKey, e.reqPath)
}

func (e *missingDlmErr) IsPublic() bool {
	return false
}

func (e *missingDlmErr) Panic() {
	panic(e)
}

func PanicIf(e error) {
	if e != nil {
		if pErr, ok := e.(PermissionedError); !ok {
			(&externalError{
				AppError:      *externalAppErr,
				OriginalError: e,
			}).Panic()
		} else {
			pErr.Panic()
		}
	}
}

func IsSqlErrNoRowsElsePanicIf(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	PanicIf(err)
	return false
}

func FmtPanic(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a...))
}
