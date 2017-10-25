package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Error(t *testing.T) {
	e := &AppError{Code: "test", Message: "yo ho ho"}
	assert.Equal(t, "code:\n\ttest\nmessage:\n\tyo ho ho\n", e.Error())
}
