package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ValidateStringParam_tooShort(t *testing.T) {
	err := ValidateStringParam("test", "yo", 3, 5, []string{"321"})
	assert.Equal(t, "test must be between 3 and 5 utf8 characters long and match all regexs [321]", err.Error())
}

func Test_ValidateStringParam_tooLong(t *testing.T) {
	err := ValidateStringParam("test", "ooppaa", 3, 5, []string{"321"})
	assert.NotNil(t, err)
}

func Test_ValidateStringParam_doesntMatchRegex(t *testing.T) {
	err := ValidateStringParam("test", "123", 3, 5, []string{"321"})
	assert.NotNil(t, err)
}

func Test_ValidateStringParam_success(t *testing.T) {
	err := ValidateStringParam("test", "321", 3, 5, []string{"321"})
	assert.Nil(t, err)
}

func Test_ValidateEmail_error(t *testing.T) {
	err := ValidateEmail("fail")
	assert.NotNil(t, err)
}

func Test_ValidateEmail_success(t *testing.T) {
	err := ValidateEmail("fail@fuck.you")
	assert.Nil(t, err)
}
