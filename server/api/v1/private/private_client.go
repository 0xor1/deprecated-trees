package private

import (
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/private"
	"fmt"
)

func NewTestClient(testServerBaseUrl string) private.V1Client {
	return &testClient{
		testServerBaseUrl: testServerBaseUrl,
	}
}

type testClient struct {
	testServerBaseUrl string
}

func (c *testClient) CreateAccount(region cnst.Region, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error) {
	return _createAccount(c.testServerBaseUrl, region, account, me, myName, myDisplayName, hasAvatar)
}

func (c *testClient) DeleteAccount(region cnst.Region, shard int, account, me id.Id) error {
	return _deleteAccount(c.testServerBaseUrl, region, shard, account, me)
}

func (c *testClient) AddMembers(region cnst.Region, shard int, account, me id.Id, members []*private.AddMember) error {
	return _addMembers(c.testServerBaseUrl, region, shard, account, me, members)
}

func (c *testClient) RemoveMembers(region cnst.Region, shard int, account, me id.Id, members []id.Id) error {
	return _removeMembers(c.testServerBaseUrl, region, shard, account, me, members)
}

func (c *testClient) MemberIsOnlyAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error) {
	return _memberIsOnlyAccountOwner(c.testServerBaseUrl, region, shard, account, me)
}

func (c *testClient) SetMemberName(region cnst.Region, shard int, account, me id.Id, newName string) error {
	return _setMemberName(c.testServerBaseUrl, region, shard, account, me, newName)
}

func (c *testClient) SetMemberDisplayName(region cnst.Region, shard int, account, me id.Id, newDisplayName *string) error {
	return _setMemberDisplayName(c.testServerBaseUrl, region, shard, account, me, newDisplayName)
}

func (c *testClient) SetMemberHasAvatar(region cnst.Region, shard int, account, me id.Id, hasAvatar bool) error {
	return _setMemberHasAvatar(c.testServerBaseUrl, region, shard, account, me, hasAvatar)
}

func (c *testClient) MemberIsAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error) {
	return _memberIsAccountOwner(c.testServerBaseUrl, region, shard, account, me)
}

func NewClient(env cnst.Env, scheme, nakedHost string) private.V1Client {
	return &client{
		env:       env,
		scheme:    scheme,
		nakedHost: nakedHost,
	}
}

type client struct {
	env       cnst.Env
	scheme    string
	nakedHost string
}

func (c *client) getBaseUrl(region cnst.Region) string {
	return fmt.Sprintf("%s%s-%s-api.%s", c.scheme, c.env, region, c.nakedHost)
}

func (c *client) CreateAccount(region cnst.Region, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error) {
	return _createAccount(c.getBaseUrl(region), region, account, me, myName, myDisplayName, hasAvatar)
}

func (c *client) DeleteAccount(region cnst.Region, shard int, account, me id.Id) error {
	return _deleteAccount(c.getBaseUrl(region), region, shard, account, me)
}

func (c *client) AddMembers(region cnst.Region, shard int, account, me id.Id, members []*private.AddMember) error {
	return _addMembers(c.getBaseUrl(region), region, shard, account, me, members)
}

func (c *client) RemoveMembers(region cnst.Region, shard int, account, me id.Id, members []id.Id) error {
	return _removeMembers(c.getBaseUrl(region), region, shard, account, me, members)
}

func (c *client) MemberIsOnlyAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error) {
	return _memberIsOnlyAccountOwner(c.getBaseUrl(region), region, shard, account, me)
}

func (c *client) SetMemberName(region cnst.Region, shard int, account, me id.Id, newName string) error {
	return _setMemberName(c.getBaseUrl(region), region, shard, account, me, newName)
}

func (c *client) SetMemberDisplayName(region cnst.Region, shard int, account, me id.Id, newDisplayName *string) error {
	return _setMemberDisplayName(c.getBaseUrl(region), region, shard, account, me, newDisplayName)
}

func (c *client) SetMemberHasAvatar(region cnst.Region, shard int, account, me id.Id, hasAvatar bool) error {
	return _setMemberHasAvatar(c.getBaseUrl(region), region, shard, account, me, hasAvatar)
}

func (c *client) MemberIsAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error) {
	return _memberIsAccountOwner(c.getBaseUrl(region), region, shard, account, me)
}

func _createAccount(baseUrl string, region cnst.Region, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error) {
	respVal := 0
	val, e := createAccount.DoRequest(nil, baseUrl, region, &createAccountArgs{
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

func _deleteAccount(baseUrl string, region cnst.Region, shard int, account, me id.Id) error {
	_, e := deleteAccount.DoRequest(nil, baseUrl, region, &deleteAccountArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, nil)
	return e
}

func _addMembers(baseUrl string, region cnst.Region, shard int, account, me id.Id, members []*private.AddMember) error {
	_, e := addMembers.DoRequest(nil, baseUrl, region, &addMembersArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		Members: members,
	}, nil, nil)
	return e
}

func _removeMembers(baseUrl string, region cnst.Region, shard int, account, me id.Id, members []id.Id) error {
	_, err := removeMembers.DoRequest(nil, baseUrl, region, &removeMembersArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		Members: members,
	}, nil, nil)
	return err
}

func _memberIsOnlyAccountOwner(baseUrl string, region cnst.Region, shard int, account, me id.Id) (bool, error) {
	respVal := false
	val, e := memberIsOnlyAccountOwner.DoRequest(nil, baseUrl, region, &memberIsOnlyAccountOwnerArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}

func _setMemberName(baseUrl string, region cnst.Region, shard int, account, me id.Id, newName string) error {
	_, e := setMemberName.DoRequest(nil, baseUrl, region, &setMemberNameArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
		NewName: newName,
	}, nil, nil)
	return e
}

func _setMemberDisplayName(baseUrl string, region cnst.Region, shard int, account, me id.Id, newDisplayName *string) error {
	_, e := setMemberDisplayName.DoRequest(nil, baseUrl, region, &setMemberDisplayNameArgs{
		Shard:          shard,
		Account:        account,
		Me:             me,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func _setMemberHasAvatar(baseUrl string, region cnst.Region, shard int, account, me id.Id, hasAvatar bool) error {
	_, e := setMemberHasAvatar.DoRequest(nil, baseUrl, region, &setMemberHasAvatarArgs{
		Shard:     shard,
		Account:   account,
		Me:        me,
		HasAvatar: hasAvatar,
	}, nil, nil)
	return e
}

func _memberIsAccountOwner(baseUrl string, region cnst.Region, shard int, account, me id.Id) (bool, error) {
	respVal := false
	val, e := memberIsAccountOwner.DoRequest(nil, baseUrl, region, &memberIsAccountOwnerArgs{
		Shard:   shard,
		Account: account,
		Me:      me,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}
