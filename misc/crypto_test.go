package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bytes(t *testing.T) {
	l := 3
	ch := NewCryptoHelper()
	bs := ch.Bytes(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = ch.Bytes(l)
	assert.Equal(t, l, len(bs))
}

func Test_UrlSafeString(t *testing.T) {
	l := 3
	ch := NewCryptoHelper()
	bs := ch.UrlSafeString(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = ch.UrlSafeString(l)
	assert.Equal(t, l, len(bs))
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	ch := NewCryptoHelper()
	pwd := ch.Bytes(l)
	salt := ch.Bytes(l)
	scryptPwd := ch.ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	l = 8
	pwd = ch.Bytes(l)
	salt = ch.Bytes(l)
	scryptPwd = ch.ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
}
