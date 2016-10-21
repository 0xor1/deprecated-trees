package main

import(
	"github.com/robsix/golog"
	"github.com/jmcvetta/neoism"
	"fmt"
	"bitbucket.org/robsix/core"
	"bitbucket.org/robsix/core/user"
	"encoding/json"
)

func main(){
	log := golog.NewLog(golog.Info, "15:04:05.000", 20)
	user.NewStore(nil, nil, nil, nil, nil, nil)
	log.Info("starting")

	e := &core.Entity{Id: "yo ho ho"}
	d, _ := json.Marshal(e)

	log.Info("%s", d)

	e2 := &core.Entity{}
	json.Unmarshal(d, e2)

	log.Info("%s", e2)
	_, err := neoism.Connect("http://neo4j:root@localhost:7474")
	if err != nil{
		log.Error("%s", err)
	}

	log.Info("ending")
	fmt.Scanln()
}