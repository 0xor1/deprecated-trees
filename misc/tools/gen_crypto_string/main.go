package main

import (
	"bitbucket.org/0xor1/task_center/misc"
	"flag"
	"fmt"
)

func main() {
	var nTmp uint
	flag.UintVar(&nTmp, "n", 1, "number of crypto strings to generate")
	var lTmp uint
	flag.UintVar(&lTmp, "l", 64, "length of each crypto string")
	flag.Parse()
	n := int(nTmp)
	l := int(lTmp)
	ch := misc.NewCryptoHelper()
	for i := 0; i < n; i++ {
		str, _ := ch.UrlSafeString(l)
		fmt.Println(str)
	}
}
