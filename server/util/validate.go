package util

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

type invalidStringArgErr struct {
	AppError
	argPurpose    string
	minRuneCount  int
	maxRuneCount  int
	regexMatchers []*regexp.Regexp
}

func (e *invalidStringArgErr) Error() string {
	return fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.argPurpose, e.minRuneCount, e.maxRuneCount, e.regexMatchers)
}

func (e *invalidStringArgErr) IsPublic() bool {
	return true
}

func (e *invalidStringArgErr) Panic() {
	panic(e)
}

func invalidStringArgErrPanic(argPurpose string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	(&invalidStringArgErr{
		AppError: AppError{
			Code:    "g_isa",
			Message: "invalid string arg",
		},
		argPurpose:    argPurpose,
		minRuneCount:  minRuneCount,
		maxRuneCount:  maxRuneCount,
		regexMatchers: append(make([]*regexp.Regexp, 0, len(regexMatchers)), regexMatchers...),
	}).Panic()
}

func ValidateStringArg(argPurpose, arg string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	valRuneCount := utf8.RuneCountInString(arg)
	if valRuneCount < minRuneCount || valRuneCount > maxRuneCount {
		invalidStringArgErrPanic(argPurpose, minRuneCount, maxRuneCount, regexMatchers)
	}
	for _, regex := range regexMatchers {
		if matches := regex.MatchString(arg); !matches {
			invalidStringArgErrPanic(argPurpose, minRuneCount, maxRuneCount, regexMatchers)
		}
	}
}

var emailRegex = regexp.MustCompile(`.+@.+\..+`)

func ValidateEmail(email string) {
	ValidateStringArg("email", email, 6, 254, []*regexp.Regexp{emailRegex})
}

func ValidateLimit(limit, maxLimit int) int {
	if limit < 1 || limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func ValidateEntityCount(entityCount, maxLimit int) {
	if entityCount < 1 || entityCount > maxLimit {
		InvalidEntityCountErr.Panic()
	}
}

func ValidateExists(exists bool) {
	if !exists {
		NoSuchEntityErr.Panic()
	}
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

func ValidateMemberIsAProjectMemberWithWriteAccess(projectRole *ProjectRole) {
	if projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter) {
		InsufficientPermissionErr.Panic()
	}
}

func ValidateMemberHasProjectReadAccess(accountRole *AccountRole, projectRole *ProjectRole, projectIsPublic *bool) {
	if projectIsPublic == nil || (!*projectIsPublic && (accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter && *projectRole != ProjectReader))))) {
		InsufficientPermissionErr.Panic()
	}
}
