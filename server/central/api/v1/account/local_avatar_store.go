package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type localAvatarStore struct {
	mtx        *sync.Mutex
	absDirPath string
}

func (s *localAvatarStore) put(key string, mimeType string, data io.Reader) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	avatarBytes, err := ioutil.ReadAll(data)
	PanicIf(err)
	PanicIf(ioutil.WriteFile(path.Join(s.absDirPath, key), avatarBytes, os.ModePerm))
}

func (s *localAvatarStore) delete(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	PanicIf(os.Remove(path.Join(s.absDirPath, key)))
}

func (s *localAvatarStore) deleteAll() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	os.RemoveAll(s.absDirPath)
}
