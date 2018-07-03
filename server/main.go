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
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
)

func main() {
	SR := static.Config("config.json", private.NewClient)
	endPointSets := make([][]*endpoint.Endpoint, 0, 100)
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
	appServer := server.New(SR, endPointSets...)
	if SR.Env == cnst.LclEnv {
		fmt.Println("server running on ", SR.BindAddress)
		SR.LogError(http.ListenAndServe(SR.BindAddress, appServer))
	} else {
		m := &autocert.Manager{
			Cache:      autocert.DirCache("acme_certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(SR.AllHosts...),
		}
		go http.ListenAndServe(":http", m.HTTPHandler(nil))
		s := &http.Server{
			Addr:      ":https",
			Handler:   appServer,
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		fmt.Println("server running on autocert settings")
		SR.LogError(s.ListenAndServeTLS("", ""))
	}
}
