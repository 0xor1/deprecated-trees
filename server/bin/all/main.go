package main

import (
	"bitbucket.org/0xor1/task/server/central/api/v1/centralaccount"
	"bitbucket.org/0xor1/task/server/regional/api/v1/account"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	"bitbucket.org/0xor1/task/server/regional/api/v1/task"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"fmt"
	"net/http"
)

func main() {
	SR := static.Config("config", ".", private.NewClient)
	fmt.Println("server running on ", SR.ServerAddress)
	SR.LogError(http.ListenAndServe(SR.ServerAddress, server.New(SR, centralaccount.Endpoints, private.Endpoints, account.Endpoints, project.Endpoints, task.Endpoints)))
}
