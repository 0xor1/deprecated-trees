package main

import (
	"bitbucket.org/0xor1/task/server/util"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
)

func main() {
	fs := flag.NewFlagSet("util", flag.ExitOnError)
	var t string
	fs.StringVar(&t, "t", "b", "b for base64 bytes array or s for ASCII string")
	var nTmp uint
	fs.UintVar(&nTmp, "n", 1, "number of crypt bytes or ASCII characters to generate")
	var lTmp uint
	fs.UintVar(&lTmp, "l", 64, "length of each crypt byte array or ASCII string")
	fs.Parse(os.Args[1:])
	n := int(nTmp)
	l := int(lTmp)
	if t == "s" {
		for i := 0; i < n; i++ {
			fmt.Println(util.CryptUrlSafeString(l))
		}
	} else {
		for i := 0; i < n; i++ {
			fmt.Println(fmt.Sprintf("%s", base64.StdEncoding.EncodeToString(util.CryptBytes(l))))
		}
	}
}
