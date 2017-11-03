package misc

import (
	"database/sql"
	"fmt"
)

var (
	NotImplementedErr         = &AppError{Code: "g_ni", Message: "not implemented", Public: true}
	InvalidArgumentsErr       = &AppError{Code: "g_ia", Message: "invalid arguments", Public: true}
	idGenerationErr           = &AppError{Code: "g_ig", Message: "failed to generate id", Public: false}
	InsufficientPermissionErr = &AppError{Code: "g_ip", Message: "insufficient permissions", Public: true}
	InvalidOperationErr       = &AppError{Code: "g_io", Message: "invalid operation", Public: true}
	InvalidEntityCountErr     = &AppError{Code: "g_iece", Message: "invalid entity count", Public: true}
	externalAppErr            = &AppError{Code: "g_ea", Message: "external error occured", Public: false}
)

type PermissionedError interface {
	error
	IsPublic() bool
}

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Public  bool   `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code:\n\t%s\nmessage:\n\t%s\n", e.Code, e.Message)
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
	return fmt.Sprintf("code:\n\t%s\nmessage:\n\t%s\noriginal error:\n\t%s\n", e.Code, e.Message, e.OriginalError.Error())
}

func PanicIf(e error) {
	if e != nil {
		panic(&externalError{
			AppError:      *externalAppErr,
			OriginalError: e,
		})
	}
}

func IsSqlErrNoRowsAndPanicIf(e error) bool {
	if e == sql.ErrNoRows {
		return true
	}
	PanicIf(e)
	return false
}
