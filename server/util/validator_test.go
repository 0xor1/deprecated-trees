package util

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_ValidateStringParam_tooShort(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.Equal(t, "test must be between 3 and 5 utf8 characters long and match all regexs [321]", err.Error())
	}()
	ValidateStringParam("test", "yo", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_tooLong(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.NotNil(t, err)
	}()
	ValidateStringParam("test", "ooppaa", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_doesntMatchRegex(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.NotNil(t, err)
	}()
	ValidateStringParam("test", "123", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateStringParam_success(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	ValidateStringParam("test", "321", 3, 5, []*regexp.Regexp{regexp.MustCompile("321")})
}

func Test_ValidateEmail_error(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	ValidateEmail("fail")
}

func Test_ValidateEmail_success(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	ValidateEmail("fail@fuck.you")
}
