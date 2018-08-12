package project

import (
	"github.com/0xor1/trees/server/api/v1/account"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/systemtest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_system(t *testing.T) {
	systemtest.Run(t, func(base *systemtest.Base) {
		accountClient := account.NewClient(base.TestServerURL)
		client := NewClient(base.TestServerURL)

		accountClient.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, account.Fields{PublicProjectsEnabled: &field.Bool{true}})

		p1Desc := "p1_desc"
		p2Desc := "p2_desc"
		p3Desc := "p3_desc"
		proj, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, "a-p1", &p1Desc, 8, 5, nil, nil, true, false, nil)
		assert.Nil(t, err)
		proj2, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, "b-p2", &p2Desc, 8, 5, nil, nil, true, false, nil)
		proj3, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, "c-p3", &p3Desc, 8, 5, nil, nil, true, false, nil)
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, Fields{
			IsPublic:    &field.Bool{true},
			HoursPerDay: &field.UInt8{6},
			DaysPerWeek: &field.UInt8{6},
		})
		proj, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id)
		assert.Equal(t, "a-p1", proj.Name)
		assert.Equal(t, "p1_desc", *proj.Description)
		assert.Equal(t, true, proj.IsParallel)
		assert.Equal(t, true, proj.IsPublic)
		projRes, err := client.GetSet(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, nil, nil, nil, nil, nil, false, cnst.SortByCreatedOn, true, nil, 1)
		assert.Equal(t, 1, len(projRes.Projects))
		assert.True(t, projRes.More)
		assert.Equal(t, proj.Name, projRes.Projects[0].Name)
		projRes, err = client.GetSet(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, nil, nil, nil, nil, nil, false, cnst.SortByCreatedOn, true, &proj.Id, 100)
		assert.Equal(t, 2, len(projRes.Projects))
		assert.False(t, projRes.More)
		assert.Equal(t, proj2.Name, projRes.Projects[0].Name)
		assert.Equal(t, proj3.Name, projRes.Projects[1].Name)
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, Fields{IsArchived: &field.Bool{true}})
		projRes, err = client.GetSet(base.Ali.CSS, base.Region, 0, base.Org.Id, nil, nil, nil, nil, nil, nil, nil, true, cnst.SortByCreatedOn, true, nil, 100)
		assert.Equal(t, 1, len(projRes.Projects))
		assert.False(t, projRes.More)
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, Fields{IsArchived: &field.Bool{false}})
		aliP := &AddProjectMember{}
		aliP.Id = base.Ali.Info.Me.Id
		aliP.Role = cnst.ProjectAdmin
		bobP := &AddProjectMember{}
		bobP.Id = base.Bob.Info.Me.Id
		bobP.Role = cnst.ProjectWriter
		catP := &AddProjectMember{}
		catP.Id = base.Cat.Info.Me.Id
		catP.Role = cnst.ProjectReader
		client.AddMembers(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, []*AddProjectMember{aliP, bobP, catP})
		accountClient.SetMemberRole(base.Ali.CSS, base.Region, 0, base.Org.Id, bobP.Id, cnst.AccountMemberOfOnlySpecificProjects)
		client.SetMemberRole(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, bobP.Id, cnst.ProjectReader)
		memRes, err := client.GetMembers(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, 100)
		assert.Equal(t, 3, len(memRes.Members))
		assert.False(t, memRes.More)
		assert.True(t, memRes.Members[0].Id.Equal(base.Ali.Info.Me.Id))
		assert.True(t, memRes.Members[1].Id.Equal(base.Bob.Info.Me.Id))
		assert.True(t, memRes.Members[2].Id.Equal(base.Cat.Info.Me.Id))
		bobMe, err := client.GetMe(base.Bob.CSS, base.Region, 0, base.Org.Id, proj.Id)
		assert.True(t, bobMe.Id.Equal(base.Bob.Info.Me.Id))
		activities, err := client.GetActivities(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, nil, 100)
		assert.Equal(t, 10, len(activities))
		client.RemoveMembers(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, []id.Id{base.Bob.Info.Me.Id, base.Cat.Info.Me.Id})
		client.Delete(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id)
	}, account.Endpoints, Endpoints)
}
