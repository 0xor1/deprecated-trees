package cnst

import (
	"bitbucket.org/0xor1/trees/server/util/err"
	"github.com/0xor1/panic"
	"strconv"
	"strings"
)

const (
	LclEnv = Env("lcl")
	DevEnv = Env("dev")
	StgEnv = Env("stg")
	ProEnv = Env("pro")

	CentralRegion = Region("central")
	USWRegion     = Region("usw")
	USERegion     = Region("use")
	EUWRegion     = Region("euw")
	ASPRegion     = Region("asp")
	AUSRegion     = Region("aus")

	LightTheme      = Theme(0)
	DarkTheme       = Theme(1)
	ColorBlindTheme = Theme(2)

	AccountOwner                        = AccountRole(0)
	AccountAdmin                        = AccountRole(1)
	AccountMemberOfAllProjects          = AccountRole(2)
	AccountMemberOfOnlySpecificProjects = AccountRole(3)

	ProjectAdmin  = ProjectRole(0)
	ProjectWriter = ProjectRole(1)
	ProjectReader = ProjectRole(2)

	SortByName        = SortBy("name")
	SortByCreatedOn   = SortBy("createdon")
	SortByDisplayName = SortBy("displayname")
	SortByStartOn     = SortBy("starton")
	SortByDueOn       = SortBy("dueon")
)

var (
	invalidConstantValueErr = &err.Err{Code: "u_c_icv", Message: "invalid constant value"}
)

type Env string

func (e *Env) Validate() {
	panic.IfTrue(e != nil && !(*e == LclEnv || *e == DevEnv || *e == StgEnv || *e == ProEnv), invalidConstantValueErr)
}

func (e *Env) String() string {
	return string(*e)
}

func (e *Env) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), `"`)
	*e = Env(val)
	e.Validate()
	return nil
}

type Region string

func (r *Region) Validate() {
	panic.IfTrue(r != nil && !(*r == CentralRegion || *r == USWRegion || *r == USERegion || *r == EUWRegion || *r == ASPRegion || *r == AUSRegion), invalidConstantValueErr)
}

func (r *Region) ValidateForDataRegions() {
	panic.IfTrue(r != nil && !(*r == USWRegion || *r == USERegion || *r == EUWRegion || *r == ASPRegion || *r == AUSRegion), invalidConstantValueErr)
}

func (r *Region) String() string {
	return string(*r)
}

func (r *Region) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), `"`)
	*r = Region(val)
	r.Validate()
	return nil
}

type Theme uint8

func (t *Theme) Validate() {
	panic.IfTrue(t != nil && !(*t == LightTheme || *t == DarkTheme || *t == ColorBlindTheme), invalidConstantValueErr)
}

func (t *Theme) String() string {
	if t == nil {
		return ""
	}
	return strconv.Itoa(int(*t))
}

func (t *Theme) UnmarshalJSON(raw []byte) error {
	val, err := strconv.ParseUint(string(raw), 10, 8)
	if err != nil {
		return err
	}
	*t = Theme(val)
	t.Validate()
	return nil
}

type AccountRole uint8

func (r *AccountRole) Validate() {
	panic.IfTrue(r != nil && !(*r == AccountOwner || *r == AccountAdmin || *r == AccountMemberOfAllProjects || *r == AccountMemberOfOnlySpecificProjects), invalidConstantValueErr)
}

func (r *AccountRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *AccountRole) UnmarshalJSON(raw []byte) error {
	val, e := strconv.ParseUint(string(raw), 10, 8)
	panic.If(e)
	*r = AccountRole(val)
	r.Validate()
	return nil
}

type ProjectRole uint8

func (r *ProjectRole) Validate() {
	panic.IfTrue(r != nil && !(*r == ProjectAdmin || *r == ProjectWriter || *r == ProjectReader), invalidConstantValueErr)
}

func (r *ProjectRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *ProjectRole) UnmarshalJSON(raw []byte) error {
	val, e := strconv.ParseUint(string(raw), 10, 8)
	panic.If(e)
	*r = ProjectRole(val)
	r.Validate()
	return nil
}

type SortBy string

func (sb *SortBy) Validate() {
	panic.IfTrue(sb != nil && !(*sb == SortByName || *sb == SortByDisplayName || *sb == SortByCreatedOn || *sb == SortByStartOn || *sb == SortByDueOn), invalidConstantValueErr)
}

func (sb *SortBy) String() string {
	return string(*sb)
}

func (sb *SortBy) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), `"`)
	*sb = SortBy(val)
	sb.Validate()
	return nil
}
