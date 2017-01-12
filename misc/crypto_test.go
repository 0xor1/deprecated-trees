package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenerateCryptoBytes(t *testing.T) {
	l := 3
	bs, err := CryptoBytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = CryptoBytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_GenerateCryptoUrlSafeString(t *testing.T) {
	l := 3
	bs, err := CryptoUrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = CryptoUrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	pwd, _ := CryptoBytes(l)
	salt, _ := CryptoBytes(l)
	scryptPwd, err := ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
	l = 8
	pwd, _ = CryptoBytes(l)
	salt, _ = CryptoBytes(l)
	scryptPwd, err = ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
}
