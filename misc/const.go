package misc

import (
	"strconv"
	"strings"
)

const (
	LightTheme      = Theme(0)
	DarkTheme       = Theme(1)
	ColorBlindTheme = Theme(2)

	OrgOwner                        = OrgRole(0)
	OrgAdmin                        = OrgRole(1)
	OrgMemberOfAllProjects          = OrgRole(2)
	OrgMemberOfOnlySpecificProjects = OrgRole(3)

	ProjectAdmin  = ProjectRole(0)
	ProjectWriter = ProjectRole(1)
	ProjectReader = ProjectRole(2)

	SortAsc  = SortDirection("asc")
	SortDesc = SortDirection("desc")
)

var (
	invalidConstantValueErr = &Error{Code: 104, Msg: "invalid constant value", IsPublic: true}
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

type OrgRole uint8

func (r *OrgRole) Validate() {
	if r == nil || (*r != OrgOwner && *r != OrgAdmin && *r != OrgMemberOfAllProjects && *r != OrgMemberOfOnlySpecificProjects) {
		*r = OrgMemberOfOnlySpecificProjects
		panic(invalidConstantValueErr)
	}
}

func (r *OrgRole) UnmarshalJSON(raw []byte) error {
	val, err := strconv.ParseUint(string(raw), 10, 8)
	if err != nil {
		return err
	}
	*r = OrgRole(val)
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
