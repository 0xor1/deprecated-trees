package main

import(
	"github.com/uber-go/zap"
	"bitbucket.org/robsix/task_center/central_directory/user"
	"fmt"
)

func main(){
	log := zap.New(zap.NewTextEncoder(), zap.DebugLevel)
	api, _ := user.NewMemApi([]string{`^\w*$`}, []string{`[0-9]`, `[a-z]`, `[A-Z]`, `[\W]`}, 3, 100, 1, 30, 8, 20, 40, 128, 16384, 8, 1, 32, log)
	api.Register("PilotWave", "dan.rob@gmail.com", "S0m3%PwD")
	activationCode := ""
	log.Info("please enter the activation code:")
	fmt.Scan(&activationCode)
	id, err := api.Activate(activationCode)
	log.Info("main", zap.Base64("id", id), zap.Error(err))
	id, err = api.Activate(activationCode)
	log.Info("main", zap.Base64("id", id), zap.Error(err))
}
