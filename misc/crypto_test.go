package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bytes(t *testing.T) {
	l := 3
	ch := NewCryptoHelper()
	bs, err := ch.Bytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = ch.Bytes(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_UrlSafeString(t *testing.T) {
	l := 3
	ch := NewCryptoHelper()
	bs, err := ch.UrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
	l = 5
	bs, err = ch.UrlSafeString(l)
	assert.Equal(t, l, len(bs))
	assert.Nil(t, err)
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	ch := NewCryptoHelper()
	pwd, _ := ch.Bytes(l)
	salt, _ := ch.Bytes(l)
	scryptPwd, err := ch.ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
	l = 8
	pwd, _ = ch.Bytes(l)
	salt, _ = ch.Bytes(l)
	scryptPwd, err = ch.ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	assert.Nil(t, err)
}
