package err

import (
	"database/sql"
	"fmt"
	"github.com/0xor1/panic"
)

type Http struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Http) Error() string {
	return e.Message
}

func IsSqlErrNoRowsElsePanicIf(e error) bool {
	if e == sql.ErrNoRows {
		return true
	}
	panic.IfNotNil(e)
	return false
}

func HttpPanicf(condition bool, code int, messageFmt string, messageArgs ...interface{}) {
	if condition {
		panic.IfNotNil(&Http{Code: code, Message: fmt.Sprintf(messageFmt, messageArgs...)})
	}
}
