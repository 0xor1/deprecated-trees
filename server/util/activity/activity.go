package activity

import (
	"bitbucket.org/0xor1/task/server/util/id"
	"time"
)

type Activity struct {
	OccurredOn time.Time `json:"occurredOn"`
	Member     id.Id     `json:"member"`
	Item       id.Id     `json:"item"`
	ItemType   string    `json:"itemType"`
	Action     string    `json:"action"`
	ItemName   *string   `json:"itemName,omitempty"`
	ExtraInfo  *string   `json:"extraInfo,omitempty"`
}