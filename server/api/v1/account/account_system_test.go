package account

import (
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/field"
	"bitbucket.org/0xor1/trees/server/util/systemtest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_System(t *testing.T) {
	systemtest.Run(t, func(base *systemtest.Base) {
		client := NewClient(base.TestServer.URL)
		acc, err := client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id)
		assert.Nil(t, err)
		assert.False(t, acc.PublicProjectsEnabled)
		assert.Equal(t, uint8(8), acc.HoursPerDay)
		assert.Equal(t, uint8(5), acc.DaysPerWeek)
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, Fields{
			PublicProjectsEnabled: &field.Bool{true},
			HoursPerDay:           &field.UInt8{6},
			DaysPerWeek:           &field.UInt8{6},
		})
		acc, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id)
		assert.Nil(t, err)
		assert.True(t, acc.PublicProjectsEnabled)
		assert.Equal(t, uint8(6), acc.HoursPerDay)
		assert.Equal(t, uint8(6), acc.DaysPerWeek)
		client.SetMemberRole(base.Ali.CSS, base.Region, 0, base.Org.Id, base.Bob.Info.Me.Id, cnst.AccountMemberOfAllProjects)
		membersRes, err := client.GetMembers(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, nil, 2)
		assert.True(t, membersRes.More)
		assert.Equal(t, 2, len(membersRes.Members))
		assert.True(t, base.Ali.Info.Me.Id.Equal(membersRes.Members[0].Id))
		assert.True(t, base.Bob.Info.Me.Id.Equal(membersRes.Members[1].Id))
		membersRes, err = client.GetMembers(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, &membersRes.Members[0].Id, 100)
		assert.False(t, membersRes.More)
		assert.Equal(t, 3, len(membersRes.Members))
		assert.True(t, base.Bob.Info.Me.Id.Equal(membersRes.Members[0].Id))
		assert.True(t, base.Cat.Info.Me.Id.Equal(membersRes.Members[1].Id))
		activities, err := client.GetActivities(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, nil, nil, 100)
		assert.Equal(t, 8, len(activities))
		me, err := client.GetMe(base.Bob.CSS, base.Region, 0, base.Org.Id)
		assert.Equal(t, cnst.AccountMemberOfAllProjects, me.Role)
		assert.True(t, base.Bob.Info.Me.Id.Equal(me.Id))
		assert.Equal(t, true, me.IsActive)
	}, Endpoints)
}
