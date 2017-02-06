package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newEmailLinkMailer_notImplementedPanic(t *testing.T) {
	defer func(){
		err := recover().(error)
		assert.Equal(t, NotImplementedErr, err)
	}()
	newEmailLinkMailer()
}