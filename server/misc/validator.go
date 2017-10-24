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

func (e *InvalidStringParamErr) IsPrivate() bool {
	return false
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

func ValidateOffsetAndLimitParams(offset, limit, maxLimit int) (int, int) {
	if limit < 1 || limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return offset, limit
}

func ValidateMemberHasAccountOwnerAccess(accountRole *AccountRole) {
	if accountRole == nil || *accountRole != AccountOwner {
		InsufficientPermissionErr.Panic()
	}
}

func ValidateMemberHasAccountAdminAccess(accountRole *AccountRole) {
	if accountRole == nil || (*accountRole != AccountOwner && *accountRole != AccountAdmin) {
		InsufficientPermissionErr.Panic()
	}
}

func ValidateMemberHasProjectAdminAccess(accountRole *AccountRole, projectRole *ProjectRole) {
	if accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || *projectRole != ProjectAdmin)) {
		InsufficientPermissionErr.Panic()
	}
}

func ValidateMemberHasProjectWriteAccess(accountRole *AccountRole, projectRole *ProjectRole) {
	if accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter))) {
		InsufficientPermissionErr.Panic()
	}
}

func ValidateMemberHasProjectReadAccess(accountRole *AccountRole, projectRole *ProjectRole, projectIsPublic *bool) {
	if projectIsPublic == nil || (!*projectIsPublic && (accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter && *projectRole != ProjectReader))))) {
		InsufficientPermissionErr.Panic()
	}
}
