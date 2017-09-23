package main

import (
	"bitbucket.org/0xor1/task/server/misc"
	"flag"
	"fmt"
)

func main() {
	var nTmp uint
	flag.UintVar(&nTmp, "n", 1, "number of crypto byte arrays to generate")
	var lTmp uint
	flag.UintVar(&lTmp, "l", 64, "length of each crypto byte array")
	flag.Parse()
	n := int(nTmp)
	l := int(lTmp)
	ch := misc.NewCryptoHelper()
	for i := 0; i < n; i++ {
		bs, _ := ch.CryptoBytes(l)
		fmt.Println(fmt.Sprintf("%x", bs))
	}
}
