package validate

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_ValidateStringParam_tooShort(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.Equal(t, `invalid test arg, min rune count: 3 max rune count: 5`, err.Error())
	}()
	StringArg("test", "yo", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_tooLong(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.NotNil(t, err)
	}()
	StringArg("test", "ooppaa", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_doesntMatchRegex(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.NotNil(t, err)
	}()
	StringArg("test", "123", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_success(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	StringArg("test", "321", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateEmail_error(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	Email("fail")
}

func Test_ValidateEmail_success(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	Email("fail@fuck.you")
}
