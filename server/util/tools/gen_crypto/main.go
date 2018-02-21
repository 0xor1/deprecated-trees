package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
)

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	fs := flag.NewFlagSet("util", flag.ExitOnError)
	var t string
	fs.StringVar(&t, "t", "b", "b for hex bytes array or s for ASCII string")
	var nTmp uint
	fs.UintVar(&nTmp, "n", 1, "number of crypto bytes or ASCII characters to generate")
	var lTmp uint
	fs.UintVar(&lTmp, "l", 64, "length of each crypto byte array or ASCII string")
	fs.Parse(os.Args[1:])
	n := int(nTmp)
	l := int(lTmp)
	if t == "s" {
		for i := 0; i < n; i++ {
			fmt.Println(createUrlSafeString(l))
		}
	} else {
		for i := 0; i < n; i++ {
			fmt.Println(fmt.Sprintf("%x", createBytes(l)))
		}
	}
}

func createUrlSafeString(l int) string {
	buf := make([]rune, l)
	urlSafeRunesLength := big.NewInt(int64(len(urlSafeRunes)))
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, urlSafeRunesLength)
		if err != nil {
			panic(err)
		}
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf)
}

func createBytes(l int) []byte {
	k := make([]byte, l)
	_, err := io.ReadFull(rand.Reader, k)
	if err != nil {
		panic(err)
	}
	return k
}
