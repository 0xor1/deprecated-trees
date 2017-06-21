package misc

import (
	"bytes"
	"encoding/hex"
	. "github.com/pborman/uuid"
	"time"
)

const (
	OrgOwner = OrgRole(0)
	OrgAdmin = OrgRole(1)
	OrgMemberOfAllProjects = OrgRole(2)
	OrgMemberOfOnlySpecificProjects = OrgRole(3)

	ProjectAdmin = ProjectRole(0)
	ProjectWriter = ProjectRole(1)
	ProjectReader = ProjectRole(2)
)

type OrgRole uint8

type ProjectRole uint8

type Id UUID

func (id Id) String() string {
	return hex.EncodeToString(id)
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id, other)
}

func (id Id) GreaterThanOrEqualTo(other Id) bool {
	return bytes.Compare(id, other) > -1
}

func (id Id) Copy() Id {
	return Id(append(make([]byte, 0, 16), []byte(id)...))
}

type Entity struct {
	Id Id `json:"id"`
}

type GenEntity func() (*Entity, error)

func NewEntity() *Entity {
	return &Entity{
		Id: NewId(),
	}
}

type NamedEntity struct {
	Entity
	Name string `json:"name"`
}

type AddMemberInternal struct {
	NamedEntity
	Role OrgRole `json:"orgRole"`
}

type AddMemberExternal struct {
	Entity
	Role OrgRole `json:"orgRole"`
}

type GenNamedEntity func(name string) *NamedEntity

func NewNamedEntity(name string) *NamedEntity {
	return &NamedEntity{
		Entity: Entity{
			Id: NewId(),
		},
		Name: name,
	}
}

type CreatedNamedEntity struct {
	NamedEntity
	CreatedOn time.Time `json:"createdOn"`
}

type GenCreatedNamedEntity func(name string) *CreatedNamedEntity

func NewCreatedNamedEntity(name string) *CreatedNamedEntity {
	return &CreatedNamedEntity{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: NewId(),
			},
			Name: name,
		},
		CreatedOn: time.Now().UTC(),
	}
}

type GenId func() Id

//returns version 1 uuid as a byte slice
func NewId() Id {
	id := NewUUID()
	if id == nil {
		panic(idGenerationErr)
	}
	return Id(id)
}
