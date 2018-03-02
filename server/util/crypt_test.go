package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bytes(t *testing.T) {
	l := 3
	bs := CryptBytes(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = CryptBytes(l)
	assert.Equal(t, l, len(bs))
}

func Test_UrlSafeString(t *testing.T) {
	l := 3
	bs := CryptUrlSafeString(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = CryptUrlSafeString(l)
	assert.Equal(t, l, len(bs))
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	pwd := CryptBytes(l)
	salt := CryptBytes(l)
	scryptPwd := ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	l = 8
	pwd = CryptBytes(l)
	salt = CryptBytes(l)
	scryptPwd = ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
}
