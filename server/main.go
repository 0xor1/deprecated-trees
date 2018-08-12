package main

import (
	"crypto/tls"
	"fmt"
	"github.com/0xor1/trees/server/api/v1/account"
	"github.com/0xor1/trees/server/api/v1/centralaccount"
	"github.com/0xor1/trees/server/api/v1/private"
	"github.com/0xor1/trees/server/api/v1/project"
	"github.com/0xor1/trees/server/api/v1/task"
	"github.com/0xor1/trees/server/api/v1/timelog"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/endpoint"
	"github.com/0xor1/trees/server/util/server"
	"github.com/0xor1/trees/server/util/static"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"time"
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
		httpServer := &http.Server{
			Addr:              ":http",
			Handler:           m.HTTPHandler(nil),
			ReadTimeout:       10 * time.Millisecond,
			ReadHeaderTimeout: 10 * time.Millisecond,
			WriteTimeout:      10 * time.Millisecond,
		}
		go httpServer.ListenAndServe()
		httpsServer := &http.Server{
			Addr:      ":https",
			Handler:   appServer,
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		fmt.Println("server running on autocert settings")
		SR.LogError(httpsServer.ListenAndServeTLS("", ""))
	}
}
