package main

import(
	"github.com/uber-go/zap"
	"bitbucket.org/robsix/core/user"
)

func main(){
	log := zap.New(zap.NewTextEncoder())
	api := user.NewNeoApi()
	api.Activate("")
	log.Info("some shit")
}
