package misc

import (
	"strconv"
	"strings"
)

const (
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

	SortAsc  = SortDirection("asc")
	SortDesc = SortDirection("desc")
)

var (
	invalidConstantValueErr = &Error{Code: "g_icv", Msg: "invalid constant value", IsPublic: true}
)

type Theme uint8

func (t *Theme) Validate() {
	if t == nil || (*t != LightTheme && *t != DarkTheme && *t != ColorBlindTheme) {
		*t = LightTheme
		panic(invalidConstantValueErr)
	}
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
	if r == nil || (*r != AccountOwner && *r != AccountAdmin && *r != AccountMemberOfAllProjects && *r != AccountMemberOfOnlySpecificProjects) {
		*r = AccountMemberOfOnlySpecificProjects
		panic(invalidConstantValueErr)
	}
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
	if r == nil || (*r != ProjectAdmin && *r != ProjectWriter && *r != ProjectReader) {
		*r = ProjectReader
		panic(invalidConstantValueErr)
	}
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

type SortDirection string

func (sd *SortDirection) Validate() {
	if sd == nil || (*sd != SortAsc && *sd != SortDesc) {
		*sd = SortAsc
		panic(invalidConstantValueErr)
	}
}

func (sd *SortDirection) UnmarshalJSON(raw []byte) error {
	val := strings.Trim(strings.ToLower(string(raw)), " ")
	*sd = SortDirection(val)
	sd.Validate()
	return nil
}
