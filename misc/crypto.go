package misc

import (
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
)

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func CryptoBytes(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		panic(err)
	}
	return k
}

func CryptoUrlSafeString(length int) string {
	buf := make([]rune, length)
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

func ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte {
	key, err := scrypt.Key(password, salt, N, r, p, keyLen)
	if err != nil {
		panic(err)
	}
	return key
}
