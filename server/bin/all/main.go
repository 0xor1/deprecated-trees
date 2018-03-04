package main

import (
	. "bitbucket.org/0xor1/task/server/config"
	"net/http"
)

func main() {
	staticResources := Config("config", ".")
	http.ListenAndServe(staticResources.ServerAddress, staticResources)
}
