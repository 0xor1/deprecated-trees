package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenerateCryptoBytes(t *testing.T) {
	l := 3
	bs, err := GenerateCryptoBytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = GenerateCryptoBytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_GenerateCryptoUrlSafeString(t *testing.T) {
	l := 3
	bs, err := GenerateCryptoUrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = GenerateCryptoUrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	pwd, _ := GenerateCryptoBytes(l)
	salt, _ := GenerateCryptoBytes(l)
	scryptPwd, err := ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
	l = 8
	pwd, _ = GenerateCryptoBytes(l)
	salt, _ = GenerateCryptoBytes(l)
	scryptPwd, err = ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
}
