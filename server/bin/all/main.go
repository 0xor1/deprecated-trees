package main

import (
	. "bitbucket.org/0xor1/task/server/config"
)

func main() {
	staticResources := Config("config", ".")
}
