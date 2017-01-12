package misc

import (
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
)

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type CryptoHelper interface {
	Bytes(length int) ([]byte, error)
	UrlSafeString(length int) (string, error)
	ScryptKey(password, salt []byte, N, r, p, keyLen int) ([]byte, error)
}

func NewCryptoHelper() CryptoHelper {
	return &cryptoHelper{}
}

type cryptoHelper struct {
}

func (c *cryptoHelper) Bytes(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (c *cryptoHelper) UrlSafeString(length int) (string, error) {
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

func (c *cryptoHelper) ScryptKey(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	return scrypt.Key(password, salt, N, r, p, keyLen)
}
