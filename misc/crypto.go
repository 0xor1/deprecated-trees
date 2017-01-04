package misc

import (
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
)

type GenCryptoBytes func(int) ([]byte, error)

func GenerateCryptoBytes(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type GenCryptoUrlSafeString func(int) (string, error)

func GenerateCryptoUrlSafeString(length int) (string, error) {
	buf := make([]rune, length)
	urlSafeRunesLength := big.NewInt(int64(len(urlSafeRunes)))
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, urlSafeRunesLength)
		if err != nil {
			return "", err
		}
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf), nil
}

type GenScryptKey func(password, salt []byte, N, r, p, keyLen int) ([]byte, error)

func ScryptKey(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	return scrypt.Key(password, salt, N, r, p, keyLen)
}
