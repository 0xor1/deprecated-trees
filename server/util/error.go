package util

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	NotImplementedErr         = &AppError{Code: "g_ni", Message: "not implemented", Public: true}
	InvalidArgumentsErr       = &AppError{Code: "g_ia", Message: "invalid arguments", Public: true}
	idGenerationErr           = &AppError{Code: "g_ig", Message: "failed to generate id", Public: false}
	idParseErr                = &AppError{Code: "g_ip", Message: "failed to parse id", Public: false}
	InsufficientPermissionErr = &AppError{Code: "g_ip", Message: "insufficient permissions", Public: true}
	InvalidOperationErr       = &AppError{Code: "g_io", Message: "invalid operation", Public: true}
	InvalidEntityCountErr     = &AppError{Code: "g_iec", Message: "invalid entity count", Public: true}
	NoSuchEntityErr           = &AppError{Code: "g_nse", Message: "no such entity", Public: true}
	noChangeMadeErr           = &AppError{Code: "g_nc", Message: "no change made", Public: true}
	invalidRegionErr          = &AppError{Code: "g_ir", Message: "invalid region", Public: false}
	invalidEndpointErr        = &AppError{Code: "g_ie", Message: "invalid endpoint", Public: false}
	externalAppErr            = &AppError{Code: "g_ea", Message: "external error occurred", Public: false}
)

type PermissionedError interface {
	error
	IsPublic() bool
	Panic()
}

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Public  bool   `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %q message: %q", e.Code, e.Message)
}

func (e *AppError) Panic() {
	panic(e)
}

func (e *AppError) IsPublic() bool {
	return e.Public
}

type externalError struct {
	AppError
	OriginalError error `json:"-"`
}

func (e *externalError) Error() string {
	return fmt.Sprintf("code: %q message: %q original err: %q", e.Code, e.Message, e.OriginalError.Error())
}

func (e *externalError) Panic() {
	panic(e)
}

type invalidStringArgErr struct {
	ArgPurpose    string
	MinRuneCount  int
	MaxRuneCount  int
	RegexMatchers []*regexp.Regexp
}

func (e *invalidStringArgErr) Error() string {
	return fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.ArgPurpose, e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers)
}

func (e *invalidStringArgErr) IsPublic() bool {
	return true
}

func newInvalidStringArgErr(argPurpose string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) *invalidStringArgErr {
	return &invalidStringArgErr{
		ArgPurpose:    argPurpose,
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]*regexp.Regexp, 0, len(regexMatchers)), regexMatchers...),
	}
}

func validateStringArg(argPurpose, arg string, minRuneCount, maxRuneCount int, regexMatchers []*regexp.Regexp) {
	valRuneCount := utf8.RuneCountInString(arg)
	if valRuneCount < minRuneCount || valRuneCount > maxRuneCount {
		panic(newInvalidStringArgErr(argPurpose, minRuneCount, maxRuneCount, regexMatchers))
	}
	for _, regex := range regexMatchers {
		if matches := regex.MatchString(arg); !matches {
			panic(newInvalidStringArgErr(argPurpose, minRuneCount, maxRuneCount, regexMatchers))
		}
	}
}

type missingDlmErr struct {
	dlmKey  string
	reqPath string
}

func (e *missingDlmErr) Error() string {
	return fmt.Sprintf("missing dlm key %q on path %s", e.dlmKey, e.reqPath)
}

func (e *missingDlmErr) IsPublic() bool {
	return false
}

func panicIf(e error) {
	if e != nil {
		(&externalError{
			AppError:      *externalAppErr,
			OriginalError: e,
		}).Panic()
	}
}
