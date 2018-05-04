package err

import (
	"database/sql"
	"fmt"
	"github.com/0xor1/panic"
)

var (
	NoSuchRegion           = &Err{Code: "u_e_nsr", Message: "no such region"}
	NotImplemented         = &Err{Code: "u_e_ni", Message: "not implemented"}
	InvalidArguments       = &Err{Code: "u_e_ia", Message: "invalid arguments"}
	InsufficientPermission = &Err{Code: "u_e_ip", Message: "insufficient permissions"}
	InvalidOperation       = &Err{Code: "u_e_io", Message: "invalid operation"}
	InvalidEntityCount     = &Err{Code: "u_e_iec", Message: "invalid entity count"}
	NoSuchEntity           = &Err{Code: "u_e_nse", Message: "no such entity"}
	External               = &Err{Code: "u_e_e", Message: "external error occurred"}
)

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Err) Error() string {
	return fmt.Sprintf("code: %q message: %q", e.Code, e.Message)
}

func IsSqlErrNoRowsElsePanicIf(e error) bool {
	if e == sql.ErrNoRows {
		return true
	}
	panic.If(e)
	return false
}

func FmtPanic(format string, a ...interface{}) {
	panic.If(fmt.Errorf(format, a...))
}
