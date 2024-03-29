package task

import (
	"github.com/0xor1/trees/server/api/v1/account"
	"github.com/0xor1/trees/server/api/v1/project"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/systemtest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_system(t *testing.T) {
	systemtest.Run(t, func(base *systemtest.Base) {
		accountClient := account.NewClient(base.TestServerURL)
		projectClient := project.NewClient(base.TestServerURL)
		client := NewClient(base.TestServerURL)

		start := time.Now()
		end := start.Add(5 * 24 * time.Hour)
		desc := "desc"
		proj, err := projectClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, "proj", &desc, 8, 5, &start, &end, true, false, []*project.AddProjectMember{{Id: base.Ali.Info.Me.Id, Role: cnst.ProjectAdmin}, {Id: base.Bob.Info.Me.Id, Role: cnst.ProjectAdmin}, {Id: base.Cat.Info.Me.Id, Role: cnst.ProjectWriter}, {Id: base.Dan.Info.Me.Id, Role: cnst.ProjectReader}})

		oneVal := uint64(1)
		twoVal := uint64(2)
		threeVal := uint64(3)
		fourVal := uint64(4)
		falseVal := false
		trueVal := true
		//create task needs extensive testing to test every avenue of the stored procedure
		taskA, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, nil, "A", &desc, true, &falseVal, nil, nil)
		assert.Equal(t, true, taskA.IsAbstract)
		assert.Equal(t, "A", taskA.Name)
		assert.Equal(t, "desc", *taskA.Description)
		assert.InDelta(t, time.Now().Unix(), taskA.CreatedOn.Unix(), 100)
		assert.Equal(t, uint64(0), taskA.TotalRemainingTime)
		assert.Equal(t, uint64(0), taskA.TotalLoggedTime)
		assert.Equal(t, uint64(0), *taskA.MinimumRemainingTime)
		assert.Equal(t, uint64(0), taskA.LinkedFileCount)
		assert.Equal(t, uint64(0), taskA.ChatCount)
		assert.Equal(t, uint64(0), *taskA.ChildCount)
		assert.Equal(t, uint64(0), *taskA.DescendantCount)
		assert.Equal(t, false, *taskA.IsParallel)
		assert.Nil(t, taskA.Member)
		client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, &taskA.Id, "B", &desc, true, &falseVal, nil, nil)
		taskC, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, nil, "C", &desc, true, &trueVal, nil, nil)
		taskD, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, &taskA.Id, "D", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		taskE, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, "E", &desc, false, nil, &base.Ali.Info.Me.Id, &twoVal)
		taskF, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskE.Id, "F", &desc, false, nil, &base.Ali.Info.Me.Id, &oneVal)
		taskG, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, "G", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		taskH, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskE.Id, "H", &desc, false, nil, &base.Ali.Info.Me.Id, &threeVal)
		taskI, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, nil, "I", &desc, false, nil, &base.Ali.Info.Me.Id, &twoVal)
		taskJ, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, &taskI.Id, "J", &desc, false, nil, &base.Ali.Info.Me.Id, &oneVal)
		taskK, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, nil, "K", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		taskL, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, &taskI.Id, "L", &desc, true, &trueVal, nil, nil)
		taskM, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskL.Id, nil, "M", &desc, false, nil, &base.Ali.Info.Me.Id, &threeVal)

		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, Fields{Name: &field.String{"PROJ"}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, Fields{Name: &field.String{"AAA"}, Description: &field.StringPtr{nil}, IsParallel: &field.Bool{true}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, Fields{IsParallel: &field.Bool{false}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{Member: &field.IdPtr{&base.Bob.Info.Me.Id}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{Member: &field.IdPtr{&base.Cat.Info.Me.Id}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{Member: &field.IdPtr{nil}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{Member: &field.IdPtr{&base.Cat.Info.Me.Id}})
		client.Edit(base.Cat.CSS, base.Region, 0, base.Org.Id, proj.Id, taskG.Id, Fields{RemainingTime: &field.UInt64{1}})

		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{IsAbstract: &field.Bool{true}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{IsAbstract: &field.Bool{false}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{IsAbstract: &field.Bool{true}})
		client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, Fields{IsAbstract: &field.Bool{false}})

		client.Move(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskG.Id, taskA.Id, nil)
		client.Move(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskG.Id, taskA.Id, &taskK.Id)
		client.Move(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskG.Id, taskA.Id, &taskJ.Id)
		client.Move(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskG.Id, taskL.Id, &taskM.Id)

		ancestors, err := client.GetAncestors(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 100)
		assert.Equal(t, 3, len(ancestors.Ancestors))
		assert.True(t, proj.Id.Equal(ancestors.Ancestors[0].Id))
		assert.True(t, taskA.Id.Equal(ancestors.Ancestors[1].Id))
		assert.True(t, taskL.Id.Equal(ancestors.Ancestors[2].Id))

		//test setting project as public and try getting info without a session
		ancestors, err = client.GetAncestors(nil, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 100)
		assert.Nil(t, ancestors)
		assert.NotNil(t, err)
		accountClient.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, account.Fields{PublicProjectsEnabled: &field.Bool{true}})
		projectClient.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, project.Fields{IsPublic: &field.Bool{true}})
		ancestors, err = client.GetAncestors(nil, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 100)
		assert.Equal(t, 3, len(ancestors.Ancestors))
		assert.True(t, proj.Id.Equal(ancestors.Ancestors[0].Id))
		assert.True(t, taskA.Id.Equal(ancestors.Ancestors[1].Id))
		assert.True(t, taskL.Id.Equal(ancestors.Ancestors[2].Id))
		projectClient.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, project.Fields{IsPublic: &field.Bool{false}})
		ancestors, err = client.GetAncestors(nil, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 100)
		assert.Nil(t, ancestors)
		assert.NotNil(t, err)

		client.Delete(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id)
		client.Delete(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskD.Id)

		task, err := client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskH.Id)
		assert.NotNil(t, 2, task)
		res, err := client.GetChildren(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, 100)
		assert.Equal(t, 3, len(res.Children))
		res, err = client.GetChildren(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, 2)
		assert.Equal(t, 2, len(res.Children))
		res, err = client.GetChildren(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskE.Id, 100)
		assert.Equal(t, 2, len(res.Children))
		res, err = client.GetChildren(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskH.Id, 100)
		assert.Equal(t, 1, len(res.Children))
		res, err = client.GetChildren(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskF.Id, 100)
		assert.Equal(t, 0, len(res.Children))
	}, account.Endpoints, project.Endpoints, Endpoints)
}
