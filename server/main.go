package main

import (
	"bitbucket.org/0xor1/trees/server/api/v1/account"
	"bitbucket.org/0xor1/trees/server/api/v1/centralaccount"
	"bitbucket.org/0xor1/trees/server/api/v1/private"
	"bitbucket.org/0xor1/trees/server/api/v1/project"
	"bitbucket.org/0xor1/trees/server/api/v1/task"
	"bitbucket.org/0xor1/trees/server/api/v1/timelog"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/server"
	"bitbucket.org/0xor1/trees/server/util/static"
	"fmt"
	"net/http"
)

func main() {
	SR := static.Config("config.json", private.NewClient)
	endPointSets := make([][]*endpoint.Endpoint, 0, 20)
	switch SR.Env {
	case cnst.LclEnv, cnst.DevEnv: //onebox environment, all endpoints run in the same service
		endPointSets = append(endPointSets, centralaccount.Endpoints, private.Endpoints, account.Endpoints, project.Endpoints, task.Endpoints, timelog.Endpoints)
	default:
		switch SR.Region {
		case cnst.CentralRegion: //central api box, only centralAccount endpoints
			endPointSets = append(endPointSets, centralaccount.Endpoints)
		default: //regional api box, all regional endpoints required
			endPointSets = append(endPointSets, private.Endpoints, account.Endpoints, project.Endpoints, task.Endpoints, timelog.Endpoints)
		}
	}
	fmt.Println("server running on ", SR.ServerAddress)
	SR.LogError(http.ListenAndServe(SR.ServerAddress, server.New(SR, endPointSets...)))
}
