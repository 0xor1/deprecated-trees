package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"io"
)

type s3AvatarStore struct {
}

func (s *s3AvatarStore) put(key string, mimeType string, size int64, data io.Reader) {
	NotImplementedErr.Panic()
}

func (s *s3AvatarStore) delete(key string) {
	NotImplementedErr.Panic()
}

func (s *s3AvatarStore) deleteAll() {
	NotImplementedErr.Panic()
}
