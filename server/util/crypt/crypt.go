package crypt

import (
	"bitbucket.org/0xor1/task/server/util/err"
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
)

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func Bytes(length int) []byte {
	k := make([]byte, length)
	_, e := io.ReadFull(rand.Reader, k)
	err.PanicIf(e)
	return k
}

func UrlSafeString(length int) string {
	buf := make([]rune, length)
	urlSafeRunesLength := big.NewInt(int64(len(urlSafeRunes)))
	for i := range buf {
		randomIdx, e := rand.Int(rand.Reader, urlSafeRunesLength)
		err.PanicIf(e)
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf)
}

func ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte {
	key, e := scrypt.Key(password, salt, N, r, p, keyLen)
	err.PanicIf(e)
	return key
}
