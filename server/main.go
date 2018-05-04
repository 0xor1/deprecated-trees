package server

import (
	"bitbucket.org/0xor1/task/server/central/api/v1/centralaccount"
	"bitbucket.org/0xor1/task/server/regional/api/v1/account"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	"bitbucket.org/0xor1/task/server/regional/api/v1/task"
	"bitbucket.org/0xor1/task/server/regional/api/v1/timelog"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"fmt"
	"net/http"
)

func main() {
	SR := static.Config("config", ".", private.NewClient)
	endPointSets := make([][]*endpoint.Endpoint, 0, 20)
	switch SR.Region {
	case "lcl", "dev": //onebox environment, all endpoints run in the same service
		endPointSets = append(endPointSets, centralaccount.Endpoints, private.Endpoints, account.Endpoints, project.Endpoints, task.Endpoints, timelog.Endpoints)
	case "central": //central api box, only centralAccount endpoints
		endPointSets = append(endPointSets, centralaccount.Endpoints)
	default: //regional api box, all regional endpoints required
		endPointSets = append(endPointSets, private.Endpoints, account.Endpoints, project.Endpoints, task.Endpoints, timelog.Endpoints)
	}
	fmt.Println("server running on ", SR.ServerAddress)
	SR.LogError(http.ListenAndServe(SR.ServerAddress, server.New(SR, endPointSets...)))
}
