package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"io"
)

type s3AvatarStore struct {
}

func (s *s3AvatarStore) put(key string, mimeType string, size int64, data io.Reader) {
	panic(NotImplementedErr)
}

func (s *s3AvatarStore) delete(key string) {
	panic(NotImplementedErr)
}

func (s *s3AvatarStore) deleteAll() {
	panic(NotImplementedErr)
}
