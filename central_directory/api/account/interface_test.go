package user

import (
	. "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/zap"
	"testing"
)

func Test_NewApi_NilUserStoreErr(t *testing.T) {
	api, err := newApi(nil, nil, nil, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, nil)
	assert.Nil(t, api)
	assert.Equal(t, err, nilStoreErr)
}
