package validate

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/err"
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	invalidStringArgErr = &err.Err{Code: "u_v_isa", Message: "invalid string arg"}
	emailRegex          = regexp.MustCompile(`.+@.+\..+`)
)

type invalidErr struct {
	*err.Err
	ArgName       string           `json:"argName"`
	MinRuneCount  int              `json:"minRuneCount"`
	MaxRuneCount  int              `json:"maxRuneCount"`
	RegexMatchers []*regexp.Regexp `json:"regexMatchers"`
}

func (e *invalidErr) Error() string {
	return fmt.Sprintf("invalid string arg: argName: %q, minRuneCount: %d, maxRuneCount: %d, regexMatchers: %v", e.ArgName, e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers)
}

func invalidStringArgErrPanic(argName string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	panic(&invalidErr{
		Err:           invalidStringArgErr,
		ArgName:       argName,
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]*regexp.Regexp, 0, len(regexMatchers)), regexMatchers...),
	})
}

func StringArg(argPurpose, arg string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
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

func Email(email string) {
	StringArg("email", email, 6, 254, []*regexp.Regexp{emailRegex})
}

func Limit(limit, maxLimit int) int {
	if limit < 1 || limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func EntityCount(entityCount, maxLimit int) {
	if entityCount < 1 || entityCount > maxLimit {
		panic(err.InvalidEntityCount)
	}
}

func Exists(exists bool) {
	if !exists {
		panic(err.NoSuchEntity)
	}
}

func MemberHasAccountOwnerAccess(accountRole *cnst.AccountRole) {
	if accountRole == nil || *accountRole != cnst.AccountOwner {
		panic(err.InsufficientPermission)
	}
}

func MemberHasAccountAdminAccess(accountRole *cnst.AccountRole) {
	if accountRole == nil || (*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) {
		panic(err.InsufficientPermission)
	}
}

func MemberHasProjectAdminAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole) {
	if accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || *projectRole != cnst.ProjectAdmin)) {
		panic(err.InsufficientPermission)
	}
}

func MemberHasProjectWriteAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole) {
	if accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter))) {
		panic(err.InsufficientPermission)
	}
}

func MemberIsAProjectMemberWithWriteAccess(projectRole *cnst.ProjectRole) {
	if projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter) {
		panic(err.InsufficientPermission)
	}
}

func MemberHasProjectReadAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole, projectIsPublic *bool) {
	if projectIsPublic == nil || (!*projectIsPublic && (accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter && *projectRole != cnst.ProjectReader))))) {
		panic(err.InsufficientPermission)
	}
}
