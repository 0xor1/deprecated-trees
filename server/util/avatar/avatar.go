package avatar

import (
	"bitbucket.org/0xor1/task/server/util/err"
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
	if relDirPath == "" {
		panic(err.InvalidArguments)
	}
	wd, e := os.Getwd()
	err.PanicIf(e)
	absDirPath := path.Join(wd, relDirPath)
	os.MkdirAll(absDirPath, os.ModeDir)
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
	err.PanicIf(e)
	err.PanicIf(ioutil.WriteFile(path.Join(c.absDirPath, key), avatarBytes, os.ModePerm))
}

func (c *localClient) Delete(key string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	err.PanicIf(os.Remove(path.Join(c.absDirPath, key)))
}

func (c *localClient) DeleteAll() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	err.PanicIf(os.RemoveAll(c.absDirPath))
}
