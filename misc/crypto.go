package misc

import (
	"crypto/rand"
	"io"
	"math/big"
)

func GenerateCryptoBytes(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateCryptoString(length int) (string, error) {
	buf := make([]rune, length)
	urlSafeRunesLength := len(urlSafeRunes)
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, big.NewInt(int64(urlSafeRunesLength)))
		if err != nil {
			return "", err
		}
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf), nil
}
