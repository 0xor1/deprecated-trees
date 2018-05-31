package avatar

import (
	"bitbucket.org/0xor1/trees/server/util/err"
	"github.com/0xor1/panic"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type Client interface {
	MaxAvatarDim() uint
	Save(key string, mimeType string, data io.Reader)
	Delete(key string)
	DeleteAll()
}

func NewLocalClient(dir string, maxAvatarDim uint) Client {
	panic.IfTrue(dir == "", err.InvalidArguments)
	dir, e := filepath.Abs(dir)
	panic.If(e)
	return &localClient{
		mtx:          &sync.Mutex{},
		maxAvatarDim: maxAvatarDim,
		dir:          dir,
	}
}

type localClient struct {
	mtx          *sync.Mutex
	maxAvatarDim uint
	dir          string
}

func (c *localClient) MaxAvatarDim() uint {
	return c.maxAvatarDim
}

func (c *localClient) Save(key string, mimeType string, data io.Reader) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	avatarBytes, e := ioutil.ReadAll(data)
	panic.If(e)
	os.MkdirAll(c.dir, os.ModeDir)
	panic.If(ioutil.WriteFile(path.Join(c.dir, key), avatarBytes, os.ModePerm))
}

func (c *localClient) Delete(key string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	panic.If(os.Remove(path.Join(c.dir, key)))
}

func (c *localClient) DeleteAll() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	panic.If(os.RemoveAll(c.dir))
}
