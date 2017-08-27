package account

import (
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
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(path.Join(s.absDirPath, key), avatarBytes, os.ModePerm); err != nil {
		panic(err)
	}
}

func (s *localAvatarStore) delete(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if err := os.Remove(path.Join(s.absDirPath, key)); err != nil {
		panic(err)
	}

}

func (s *localAvatarStore) deleteAll() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	os.RemoveAll(s.absDirPath)
}
