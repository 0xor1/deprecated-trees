package dlm

import (
	"bitbucket.org/0xor1/task/server/util/id"
)

func ForSystem() string {
	return "sys"
}

func ForAccountMaster(accountId id.Id) string {
	return dlmKeyFor("amstr", accountId)
}

func ForAccount(accountId id.Id) string {
	return dlmKeyFor("a", accountId)
}

func ForAccountActivities(accountId id.Id) string {
	return dlmKeyFor("aa", accountId)
}

func ForAccountMember(accountId id.Id) string {
	return dlmKeyFor("am", accountId)
}

func ForAllAccountMembers(accountId id.Id) string {
	return dlmKeyFor("ams", accountId)
}

func ForProjectMaster(projectId id.Id) string {
	return dlmKeyFor("pmstr", projectId)
}

func ForProject(projectId id.Id) string {
	return dlmKeyFor("p", projectId)
}

func ForProjectActivities(projectId id.Id) string {
	return dlmKeyFor("pa", projectId)
}

func ForProjectMember(projectMemberId id.Id) string {
	return dlmKeyFor("pm", projectMemberId)
}

func ForAllProjectMembers(projectId id.Id) string {
	return dlmKeyFor("pms", projectId)
}

func ForTask(taskId id.Id) string {
	return dlmKeyFor("t", taskId)
}

func ForTasks(taskIds []id.Id) []string {
	strs := make([]string, 0, len(taskIds))
	for _, tId := range taskIds {
		strs = append(strs, ForTask(tId))
	}
	return strs
}

func dlmKeyFor(typeKey string, id id.Id) string {
	return typeKey + ":" + id.String()
}
