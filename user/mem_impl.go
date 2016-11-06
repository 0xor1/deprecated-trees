package user

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

func newMemStore() store {
	return &memStore{
		data: map[string]*fullUserInfo{},
		mtx:  &sync.RWMutex{},
	}
}

type memStore struct {
	data         map[string]*fullUserInfo
	growthFactor int
	mtx          *sync.RWMutex
}

func (s *memStore) firstWith(predicate func(u *fullUserInfo) bool) *fullUserInfo {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, u := range s.data {
		if predicate(u) {
			return cloneFullUserInfoStruct(u)
		}
	}
	return nil
}

func (s *memStore) getByUsername(username string) (*fullUserInfo, error) {
	return s.firstWith(func(u *fullUserInfo) bool {
		return u.Username == username
	}), nil
}

func (s *memStore) getByEmail(email string) (*fullUserInfo, error) {
	return s.firstWith(func(u *fullUserInfo) bool {
		return u.Email == email
	}), nil
}

func (s *memStore) getById(id string) (*fullUserInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	u := s.data[id]
	return cloneFullUserInfoStruct(u), nil
}

func (s *memStore) getByActivationCode(activationCode string) (*fullUserInfo, error) {
	return s.firstWith(func(u *fullUserInfo) bool {
		return u.ActivationCode == activationCode
	}), nil
}

func (s *memStore) getByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error) {
	return s.firstWith(func(u *fullUserInfo) bool {
		return u.NewEmailConfirmationCode == confirmationCode
	}), nil
}

func (s *memStore) getByResetPwdCode(resetPwdCode string) (*fullUserInfo, error) {
	return s.firstWith(func(u *fullUserInfo) bool {
		return u.ResetPwdCode == resetPwdCode
	}), nil
}

func (s *memStore) getByIds(ids []string) ([]*User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*User, 0, len(ids))
	for _, id := range ids {
		u := s.data[id]
		if u != nil {
			clone := u.User
			res = append(res, &clone)
		} else {
			return nil, errors.New(fmt.Sprintf("user with id %q does not exist", id))
		}
	}
	return res, nil
}

func (s *memStore) search(search string, limit int) ([]*User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*User, 0, limit)
	for _, u := range s.data {
		if strings.Contains(u.Username, search) {
			clone := u.User
			res = append(res, &clone)
		}
		if len(res) == limit {
			return res, nil
		}
	}
	return res, nil
}

func (s *memStore) create(user *fullUserInfo) error {
	if user == nil {
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.data[user.Id]; exists {
		return errors.New("user already exists")
	}

	s.data[user.Id] = cloneFullUserInfoStruct(user)

	return nil
}

func (s *memStore) update(user *fullUserInfo) error {
	if user == nil {
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.data[user.Id]; !exists {
		return errors.New("user does not exists")
	}

	s.data[user.Id] = cloneFullUserInfoStruct(user)

	return nil
}

func (s *memStore) delete(id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.data[id]; !exists {
		return errors.New("user does not exists")
	}

	delete(s.data, id)

	return nil
}

//helper
func cloneFullUserInfoStruct(u *fullUserInfo) *fullUserInfo {
	if u == nil {
		return nil
	}
	clone := *u
	if u.ActivationTime != nil {
		activationTimeClone := *u.ActivationTime
		clone.ActivationTime = &activationTimeClone
	}
	clone.ScryptSalt = append(make([]byte, 0, len(u.ScryptSalt)), u.ScryptSalt...)
	clone.ScryptPwd = append(make([]byte, 0, len(u.ScryptPwd)), u.ScryptPwd...)
	return &clone
}
