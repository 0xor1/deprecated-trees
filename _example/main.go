package main

import(
	"github.com/uber-go/zap"
	"bitbucket.org/robsix/core/user"
	"bitbucket.org/robsix/core/helper"
)

func main(){
	log := zap.New(zap.NewTextEncoder())
	linkMailer, _ := helper.NewLogLinkMailer(log)
	api, _ := user.NewMemApi(linkMailer, []string{`^\w*$`}, []string{`[0-9]`, `[a-z]`, `[A-Z]`, `[\W]`}, 3, 100, 1, 20, 8, 20, 20, 128, 16384, 8, 1, 256, log)
	api.Register("PilotWaveMedium", "dan.rob@gmail.com", "S0m3%PwD")
}
