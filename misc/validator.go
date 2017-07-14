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
	RegexMatchers []*regexp.Regexp
}

func (e *InvalidStringParamErr) Error() string {
	return fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.ParamPurpose, e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers)
}

func newInvalidStringParamErr(paramPurpose string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) *InvalidStringParamErr {
	return &InvalidStringParamErr{
		ParamPurpose:  paramPurpose,
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]*regexp.Regexp, 0, len(regexMatchers)), regexMatchers...),
	}
}

func ValidateStringParam(paramPurpose, param string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	valRuneCount := utf8.RuneCountInString(param)
	if valRuneCount < minRuneCount || valRuneCount > maxRuneCount {
		panic(newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers))
	}
	for _, regex := range regexMatchers {
		if matches := regex.MatchString(param); !matches {
			panic(newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers))
		}
	}
}

var emailRegex = regexp.MustCompile(`.+@.+\..+`)

func ValidateEmail(email string) {
	ValidateStringParam("email", email, 6, 254, []*regexp.Regexp{emailRegex})
}

func IsPrivate() bool {
	return false
}

func ValidateOffsetAndLimitParams(offset, limit, maxLimit int) (int, int) {
	if limit < 1 || limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return offset, limit
}