package http

import (
	"bitbucket.org/0xor1/task/server/central/api/v1/account"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Route(router httprouter.Router, apiV1Account account.Api) {
	router.GET("/api/v1/account/getRegions", getRegions)
}
