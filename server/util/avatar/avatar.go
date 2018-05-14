package avatar

import (
	"bitbucket.org/0xor1/task/server/util/err"
	"github.com/0xor1/panic"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type Client interface {
	MaxAvatarDim() uint
	Save(key string, mimeType string, data io.Reader)
	Delete(key string)
	DeleteAll()
}

func NewLocalClient(relDirPath string, maxAvatarDim uint) Client {
	panic.IfTrueWith(relDirPath == "", err.InvalidArguments)
	wd, e := os.Getwd()
	panic.If(e)
	absDirPath := path.Join(wd, relDirPath)
	return &localClient{
		mtx:          &sync.Mutex{},
		maxAvatarDim: maxAvatarDim,
		absDirPath:   absDirPath,
	}
}

type localClient struct {
	mtx          *sync.Mutex
	maxAvatarDim uint
	absDirPath   string
}

func (c *localClient) MaxAvatarDim() uint {
	return c.maxAvatarDim
}

func (c *localClient) Save(key string, mimeType string, data io.Reader) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	avatarBytes, e := ioutil.ReadAll(data)
	panic.If(e)
	os.MkdirAll(c.absDirPath, os.ModeDir)
	panic.If(ioutil.WriteFile(path.Join(c.absDirPath, key), avatarBytes, os.ModePerm))
}

func (c *localClient) Delete(key string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	panic.If(os.Remove(path.Join(c.absDirPath, key)))
}

func (c *localClient) DeleteAll() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	panic.If(os.RemoveAll(c.absDirPath))
}
