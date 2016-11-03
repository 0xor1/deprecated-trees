package main

import(
	"github.com/uber-go/zap"
	"bitbucket.org/robsix/core/user"
)

func main(){
	log := zap.New(zap.NewTextEncoder())
	api, _ := user.NewNeoApi(nil, nil, nil, nil, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, nil)
	api.Activate("")
	log.Info("some shit")
}
