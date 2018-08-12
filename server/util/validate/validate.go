package validate

import (
	"github.com/0xor1/panic"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/err"
	"net/http"
	"regexp"
	"unicode/utf8"
)

var (
	emailRegex = regexp.MustCompile(`.+@.+\..+`)
)

func HoursPerDay(hoursPerDay uint8) {
	panic.If(hoursPerDay == 0 || hoursPerDay > 24, "invalid hoursPerDay must be > 0 and <= 24")
}

func DaysPerWeek(daysPerWeek uint8) {
	panic.If(daysPerWeek == 0 || daysPerWeek > 7, "invalid daysPerWeek must be > 0 and <= 7")
}

func StringArg(argPurpose, arg string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	valRuneCount := utf8.RuneCountInString(arg)
	err.HttpPanicf(valRuneCount < minRuneCount || valRuneCount > maxRuneCount, http.StatusBadRequest, "invalid %s arg, min rune count: %d max rune count: %d", argPurpose, minRuneCount, maxRuneCount)
	for _, regex := range regexMatchers {
		err.HttpPanicf(!regex.MatchString(arg), http.StatusBadRequest, "invalid %s arg, regex: %v", argPurpose, regex.String())
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
	err.HttpPanicf(entityCount < 1 || entityCount > maxLimit, http.StatusBadRequest, "invalid entity count")
}

func MemberHasAccountOwnerAccess(accountRole *cnst.AccountRole) {
	checkUnauthorized(accountRole == nil || *accountRole != cnst.AccountOwner)
}

func MemberHasAccountAdminAccess(accountRole *cnst.AccountRole) {
	checkUnauthorized(accountRole == nil || (*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin))
}

func MemberHasProjectAdminAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole) {
	checkUnauthorized(accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || *projectRole != cnst.ProjectAdmin)))
}

func MemberHasProjectWriteAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole) {
	checkUnauthorized(accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter))))
}

func MemberIsAProjectMemberWithWriteAccess(projectRole *cnst.ProjectRole) {
	checkUnauthorized(projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter))
}

func MemberHasProjectReadAccess(accountRole *cnst.AccountRole, projectRole *cnst.ProjectRole, projectIsPublic *bool) {
	checkUnauthorized(projectIsPublic == nil || (!*projectIsPublic && (accountRole == nil || ((*accountRole != cnst.AccountOwner && *accountRole != cnst.AccountAdmin) && (projectRole == nil || (*projectRole != cnst.ProjectAdmin && *projectRole != cnst.ProjectWriter && *projectRole != cnst.ProjectReader))))))
}

func checkUnauthorized(condition bool) {
	err.HttpPanicf(condition, http.StatusUnauthorized, "unauthorized")
}
