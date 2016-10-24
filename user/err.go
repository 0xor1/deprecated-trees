package user

import (
	"errors"
	"fmt"
)

var (
	IncorrectPwdErr          = errors.New("pwd incorrect")
	UserNotActivated         = errors.New("user not activated")
	EmailAlreadyInUseErr     = errors.New("email already in use")
	InvalidEmailErr          = errors.New("invalid email")
	FirstNameErr             = errors.New("firstName must be provided")
	LastNameErr              = errors.New("lastName must be provided")
	UserAlreadyActivatedErr  = errors.New("user already activated")
	EmailConfirmationCodeErr = errors.New("email confirmation code is of zero length")
	NewEmailConfirmationErr  = errors.New("new email and confirmation code do not match recod")
)

type InvalidPwdErr struct {
	MinRuneCount  int
	MaxRuneCount  int
	RegexMatchers []string
}

func newInvalidPwdErr(minRuneCount, maxRuneCount int, regexMatchers []string) *InvalidPwdErr {
	return &InvalidPwdErr{
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]string, 0, len(regexMatchers)), regexMatchers...),
	}
}

func (e *InvalidPwdErr) Error() string {
	return fmt.Sprintf(fmt.Sprintf("pwd must be between %d and %d utf8 characters long and match all regexs %v", e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers))
}
