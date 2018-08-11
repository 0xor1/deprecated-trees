package timelog

import (
	"github.com/0xor1/trees/server/api/v1/account"
	"github.com/0xor1/trees/server/api/v1/project"
	"github.com/0xor1/trees/server/api/v1/task"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/systemtest"
	ti "github.com/0xor1/trees/server/util/time"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_system(t *testing.T) {
	systemtest.Run(t, func(base *systemtest.Base) {
		projectClient := project.NewClient(base.TestServer.URL)
		taskClient := task.NewClient(base.TestServer.URL)
		client := NewClient(base.TestServer.URL)

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
		taskA, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, nil, "A", &desc, true, &falseVal, nil, nil)
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
		taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, &taskA.Id, "B", &desc, true, &falseVal, nil, nil)
		taskC, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, nil, "C", &desc, true, &trueVal, nil, nil)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, proj.Id, &taskA.Id, "D", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		taskE, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, "E", &desc, false, nil, &base.Ali.Info.Me.Id, &twoVal)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskE.Id, "F", &desc, false, nil, &base.Ali.Info.Me.Id, &oneVal)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, nil, "G", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskC.Id, &taskE.Id, "H", &desc, false, nil, &base.Ali.Info.Me.Id, &threeVal)
		taskI, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, nil, "I", &desc, false, nil, &base.Ali.Info.Me.Id, &twoVal)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, &taskI.Id, "J", &desc, false, nil, &base.Ali.Info.Me.Id, &oneVal)
		_, err = taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, nil, "K", &desc, false, nil, &base.Ali.Info.Me.Id, &fourVal)
		taskL, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskA.Id, &taskI.Id, "L", &desc, true, &trueVal, nil, nil)
		taskM, err := taskClient.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskL.Id, nil, "M", &desc, false, nil, &base.Ali.Info.Me.Id, &threeVal)

		aNote := "word up!"
		tl1, err := client.Create(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 30, &aNote)
		assert.Nil(t, err)
		assert.True(t, tl1.Project.Equal(proj.Id))
		assert.True(t, tl1.Member.Equal(base.Ali.Info.Me.Id))
		assert.True(t, tl1.Task.Equal(taskM.Id))
		assert.Equal(t, *tl1.Note, aNote)
		assert.Equal(t, uint64(30), tl1.Duration)
		assert.InDelta(t, ti.NowUnixMillis()/1000, tl1.LoggedOn.Unix(), 1)

		tl2, err := client.CreateAndSetRemainingTime(base.Bob.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id, 100, 30, &aNote)
		assert.Nil(t, err)
		assert.True(t, tl2.Project.Equal(proj.Id))
		assert.True(t, tl2.Member.Equal(base.Bob.Info.Me.Id))
		assert.True(t, tl2.Task.Equal(taskM.Id))
		assert.Equal(t, *tl2.Note, aNote)
		assert.Equal(t, uint64(30), tl2.Duration)
		assert.InDelta(t, ti.NowUnixMillis()/1000, tl2.LoggedOn.Unix(), 1)

		note := "word down?"
		err = client.Edit(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, tl2.Id, Fields{Duration: &field.UInt64{Val: 100}, Note: &field.StringPtr{Val: &note}})
		assert.Nil(t, err)

		tls, err := client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, false, nil, 100)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(tls.TimeLogs))
		assert.True(t, tls.TimeLogs[0].Id.Equal(tl2.Id))
		assert.True(t, tls.TimeLogs[1].Id.Equal(tl1.Id))

		tls, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, false, &tl2.Id, 100)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tls.TimeLogs))
		assert.True(t, tls.TimeLogs[0].Id.Equal(tl1.Id))

		tls, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, &tl2.Id, &base.Bob.Info.Me.Id, &tl2.Id, false, nil, 100)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tls.TimeLogs))
		assert.True(t, tls.TimeLogs[0].Id.Equal(tl2.Id))

		assert.Nil(t, client.Delete(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, tl1.Id))

		tls, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, false, nil, 100)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tls.TimeLogs))
		assert.True(t, tls.TimeLogs[0].Id.Equal(tl2.Id))

		assert.Nil(t, taskClient.Delete(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, taskM.Id))

		tls, err = client.Get(base.Ali.CSS, base.Region, 0, base.Org.Id, proj.Id, nil, nil, nil, false, nil, 100)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tls.TimeLogs))
		assert.Equal(t, true, tls.TimeLogs[0].TaskHasBeenDeleted)
	}, account.Endpoints, project.Endpoints, task.Endpoints, Endpoints)
}
