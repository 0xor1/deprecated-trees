package private

import (
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	"github.com/0xor1/panic"
	"strings"
)

func NewClient(regions map[string]string) private.V1Client {
	lowerRegionsMap := map[string]string{}
	for k, v := range regions {
		lowerRegionsMap[strings.ToLower(k)] = v
	}
	return &client{
		regions: lowerRegionsMap,
	}
}

type client struct {
	regions map[string]string
}

func (c *client) getHost(region string) string {
	host, exists := c.regions[strings.ToLower(region)]
	panic.IfTrueWith(!exists, err.NoSuchRegion)
	return host
}

func (c *client) GetRegions() []string {
	regions := make([]string, 0, len(c.regions))
	for r := range c.regions {
		regions = append(regions, r)
	}
	return regions
}

func (c *client) IsValidRegion(region string) bool {
	_, exists := c.regions[strings.ToLower(region)]
	return exists
}

func (c *client) CreateAccount(region string, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error) {
	respVal := 0
	val, e := createAccount.DoRequest(nil, c.getHost(region), &createAccountArgs{
		Account:       account,
		Me:            me,
		MyName:        myName,
		MyDisplayName: myDisplayName,
		HasAvatar:     hasAvatar,
	}, nil, &respVal)
	if val != nil {
		return *val.(*int), e
	}
	return 0, e
}

func (c *client) DeleteAccount(region string, shard int, account, me id.Id) error {
	_, e := deleteAccount.DoRequest(nil, c.getHost(region), &deleteAccountArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(region string, shard int, account, me id.Id, members []*private.AddMember) error {
	_, e := addMembers.DoRequest(nil, c.getHost(region), &addMembersArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		Members: members,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(region string, shard int, account, me id.Id, members []id.Id) error {
	_, err := removeMembers.DoRequest(nil, c.getHost(region), &removeMembersArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		Members: members,
	}, nil, nil)
	return err
}

func (c *client) MemberIsOnlyAccountOwner(region string, shard int, account, me id.Id) (bool, error) {
	respVal := false
	val, e := memberIsOnlyAccountOwner.DoRequest(nil, c.getHost(region), &memberIsOnlyAccountOwnerArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}

func (c *client) SetMemberName(region string, shard int, account, me id.Id, newName string) error {
	_, e := setMemberName.DoRequest(nil, c.getHost(region), &setMemberNameArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		NewName: newName,
	}, nil, nil)
	return e
}

func (c *client) SetMemberDisplayName(region string, shard int, account, me id.Id, newDisplayName *string) error {
	_, e := setMemberDisplayName.DoRequest(nil, c.getHost(region), &setMemberDisplayNameArgs{
		Shard:          shard,
		Account:        account,
		Me:             me,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func (c *client) SetMemberHasAvatar(region string, shard int, account, me id.Id, hasAvatar bool) error {
	_, e := setMemberHasAvatar.DoRequest(nil, c.getHost(region), &setMemberHasAvatarArgs{
		Shard:     shard,
		Account:   account,
		Me:        me,
		HasAvatar: hasAvatar,
	}, nil, nil)
	return e
}

func (c *client) MemberIsAccountOwner(region string, shard int, account, me id.Id) (bool, error) {
	respVal := false
	val, e := memberIsAccountOwner.DoRequest(nil, c.getHost(region), &memberIsAccountOwnerArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}
