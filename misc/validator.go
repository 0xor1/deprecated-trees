package misc

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

type InvalidStringParamErr struct {
	ParamPurpose  string
	MinRuneCount  int
	MaxRuneCount  int
	RegexMatchers []string
}

func (e *InvalidStringParamErr) Error() string {
	return fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.ParamPurpose, e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers)
}

func newInvalidStringParamErr(paramPurpose string, minRuneCount, maxRuneCount int, regexMatchers []string) *InvalidStringParamErr {
	return &InvalidStringParamErr{
		ParamPurpose:  paramPurpose,
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]string, 0, len(regexMatchers)), regexMatchers...),
	}
}

func ValidateStringParam(paramPurpose, param string, minRuneCount, maxRuneCount int, regexMatchers []string) {
	valRuneCount := utf8.RuneCountInString(param)
	if valRuneCount < minRuneCount || valRuneCount > maxRuneCount {
		panic(newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers))
	}
	for _, regex := range regexMatchers {
		if matches, err := regexp.MatchString(regex, param); !matches || err != nil {
			if err != nil {
				panic(err)
			}
			panic(newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers))
		}
	}
}

func ValidateEmail(email string) {
	ValidateStringParam("email", email, 6, 254, []string{`.+@.+\..+`})
}

func IsPrivate() bool {
	return false
}
