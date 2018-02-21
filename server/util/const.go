package util

import (
	"strconv"
	"strings"
)

const (
	GET  = "GET"
	POST = "POST"

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
	SortByDisplayName = SortBy("displayname")
	SortByCreatedOn   = SortBy("createdon")
	SortByStartOn     = SortBy("starton")
	SortByDueOn       = SortBy("dueon")

	SortDirAsc  = SortDir("asc")
	SortDirDesc = SortDir("desc")
)

var (
	invalidConstantValueErr = &AppError{Code: "g_icv", Message: "invalid constant value", Public: true}
)

type Theme uint8

func (t *Theme) Validate() {
	if t == nil || !(*t == LightTheme || *t == DarkTheme || *t == ColorBlindTheme) {
		*t = LightTheme
		invalidConstantValueErr.Panic()
	}
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
	if r == nil || !(*r == AccountOwner || *r == AccountAdmin || *r == AccountMemberOfAllProjects || *r == AccountMemberOfOnlySpecificProjects) {
		*r = AccountMemberOfOnlySpecificProjects
		invalidConstantValueErr.Panic()
	}
}

func (r *AccountRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *AccountRole) UnmarshalJSON(raw []byte) error {
	val, err := strconv.ParseUint(string(raw), 10, 8)
	if err != nil {
		return err
	}
	*r = AccountRole(val)
	r.Validate()
	return nil
}

type ProjectRole uint8

func (r *ProjectRole) Validate() {
	if r == nil || !(*r == ProjectAdmin || *r == ProjectWriter || *r == ProjectReader) {
		*r = ProjectReader
		invalidConstantValueErr.Panic()
	}
}

func (r *ProjectRole) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *ProjectRole) UnmarshalJSON(raw []byte) error {
	val, err := strconv.ParseUint(string(raw), 10, 8)
	if err != nil {
		return err
	}
	*r = ProjectRole(val)
	r.Validate()
	return nil
}

type SortDir string

func (sd *SortDir) Validate() {
	if sd == nil || !(*sd == SortDirAsc || *sd == SortDirDesc) {
		*sd = SortDirAsc
		invalidConstantValueErr.Panic()
	}
}

func (sd *SortDir) String() string {
	return string(*sd)
}

func (sd *SortDir) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), " ")
	*sd = SortDir(val)
	sd.Validate()
	return nil
}

func (sd *SortDir) GtLtSymbol() string {
	if *sd == SortDirAsc {
		return ">"
	} else {
		return "<"
	}
}

type SortBy string

func (sb *SortBy) Validate() {
	if sb == nil || !(*sb == SortByName || *sb == SortByDisplayName || *sb == SortByCreatedOn || *sb == SortByStartOn || *sb == SortByDueOn) {
		*sb = SortByName
		invalidConstantValueErr.Panic()
	}
}

func (sb *SortBy) String() string {
	return string(*sb)
}

func (sb *SortBy) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), " ")
	*sb = SortBy(val)
	sb.Validate()
	return nil
}
