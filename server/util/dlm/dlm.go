package dlm

import (
	"bitbucket.org/0xor1/task/server/util/id"
)

func ForSystem() string {
	return "sys"
}

func ForAccountMaster(account id.Id) string {
	return dlmKeyFor("amstr", account)
}

func ForAccount(account id.Id) string {
	return dlmKeyFor("a", account)
}

func ForAccountActivities(account id.Id) string {
	return dlmKeyFor("aa", account)
}

func ForAccountMember(account id.Id) string {
	return dlmKeyFor("am", account)
}

func ForAllAccountMembers(account id.Id) string {
	return dlmKeyFor("ams", account)
}

func ForProjectMaster(project id.Id) string {
	return dlmKeyFor("pmstr", project)
}

func ForProject(project id.Id) string {
	return dlmKeyFor("p", project)
}

func ForProjectActivities(project id.Id) string {
	return dlmKeyFor("pa", project)
}

func ForProjectMember(projectMember id.Id) string {
	return dlmKeyFor("pm", projectMember)
}

func ForAllProjectMembers(project id.Id) string {
	return dlmKeyFor("pms", project)
}

func ForTask(task id.Id) string {
	return dlmKeyFor("t", task)
}

func ForTasks(tasks []id.Id) []string {
	strs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		strs = append(strs, ForTask(task))
	}
	return strs
}

func dlmKeyFor(typeKey string, id id.Id) string {
	return typeKey + ":" + id.String()
}
