package cnst

import (
	"github.com/0xor1/trees/server/util/err"
	"github.com/0xor1/panic"
	"net/http"
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

type Env string

func (e *Env) Validate() {
	err.HttpPanicf(e != nil && !(*e == LclEnv || *e == DevEnv || *e == StgEnv || *e == ProEnv), http.StatusBadRequest, "invalid env")
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
	err.HttpPanicf(r != nil && !(*r == CentralRegion || *r == USWRegion || *r == USERegion || *r == EUWRegion || *r == ASPRegion || *r == AUSRegion), http.StatusBadRequest, "invalid region")
}

func (r *Region) ValidateForDataRegions() {
	err.HttpPanicf(r != nil && !(*r == USWRegion || *r == USERegion || *r == EUWRegion || *r == ASPRegion || *r == AUSRegion), http.StatusBadRequest, "invalid region")
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
	err.HttpPanicf(t != nil && !(*t == LightTheme || *t == DarkTheme || *t == ColorBlindTheme), http.StatusBadRequest, "invalid theme")
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
	err.HttpPanicf(r != nil && !(*r == AccountOwner || *r == AccountAdmin || *r == AccountMemberOfAllProjects || *r == AccountMemberOfOnlySpecificProjects), http.StatusBadRequest, "invalid account role")
}

func (r *AccountRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *AccountRole) UnmarshalJSON(raw []byte) error {
	val, e := strconv.ParseUint(string(raw), 10, 8)
	panic.IfNotNil(e)
	*r = AccountRole(val)
	r.Validate()
	return nil
}

type ProjectRole uint8

func (r *ProjectRole) Validate() {
	err.HttpPanicf(r != nil && !(*r == ProjectAdmin || *r == ProjectWriter || *r == ProjectReader), http.StatusBadRequest, "invalid project role")
}

func (r *ProjectRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *ProjectRole) UnmarshalJSON(raw []byte) error {
	val, e := strconv.ParseUint(string(raw), 10, 8)
	panic.IfNotNil(e)
	*r = ProjectRole(val)
	r.Validate()
	return nil
}

type SortBy string

func (sb *SortBy) Validate() {
	err.HttpPanicf(sb != nil && !(*sb == SortByName || *sb == SortByDisplayName || *sb == SortByCreatedOn || *sb == SortByStartOn || *sb == SortByDueOn), http.StatusBadRequest, "invalid sort by")
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
