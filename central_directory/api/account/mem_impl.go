package account

//
//import (
//	"errors"
//	"fmt"
//	. "github.com/pborman/uuid"
//	"strings"
//	"sync"
//)
//
//func newMemStore() UserStore {
//	return &memStore{
//		data: map[string]*FullUserInfo{},
//		mtx:  &sync.RWMutex{},
//	}
//}
//
//type memStore struct {
//	data         map[string]*FullUserInfo
//	growthFactor int
//	mtx          *sync.RWMutex
//}
//
//func (s *memStore) firstWith(predicate func(u *FullUserInfo) bool) *FullUserInfo {
//	s.mtx.RLock()
//	defer s.mtx.RUnlock()
//	for _, u := range s.data {
//		if predicate(u) {
//			return cloneFullUserInfoStruct(u)
//		}
//	}
//	return nil
//}
//
//func (s *memStore) getByUsername(username string) (*FullUserInfo, error) {
//	return s.firstWith(func(u *FullUserInfo) bool {
//		return u.Username == username
//	}), nil
//}
//
//func (s *memStore) getByEmail(email string) (*FullUserInfo, error) {
//	return s.firstWith(func(u *FullUserInfo) bool {
//		return u.Email == email
//	}), nil
//}
//
//func (s *memStore) getById(id UUID) (*FullUserInfo, error) {
//	s.mtx.RLock()
//	defer s.mtx.RUnlock()
//	u := s.data[id.String()]
//	return cloneFullUserInfoStruct(u), nil
//}
//
//func (s *memStore) getByActivationCode(activationCode string) (*FullUserInfo, error) {
//	return s.firstWith(func(u *FullUserInfo) bool {
//		return u.ActivationCode == activationCode
//	}), nil
//}
//
//func (s *memStore) getByNewEmailConfirmationCode(confirmationCode string) (*FullUserInfo, error) {
//	return s.firstWith(func(u *FullUserInfo) bool {
//		return u.NewEmailConfirmationCode == confirmationCode
//	}), nil
//}
//
//func (s *memStore) getByResetPwdCode(resetPwdCode string) (*FullUserInfo, error) {
//	return s.firstWith(func(u *FullUserInfo) bool {
//		return u.ResetPwdCode == resetPwdCode
//	}), nil
//}
//
//func (s *memStore) getByIds(ids []UUID) ([]*User, error) {
//	s.mtx.RLock()
//	defer s.mtx.RUnlock()
//	res := make([]*User, 0, len(ids))
//	for _, id := range ids {
//		u := s.data[id.String()]
//		if u != nil {
//			clone := u.User
//			res = append(res, &clone)
//		} else {
//			return nil, errors.New(fmt.Sprintf("user with id %q does not exist", id))
//		}
//	}
//	return res, nil
//}
//
//func (s *memStore) search(search string, limit int) ([]*User, error) {
//	s.mtx.RLock()
//	defer s.mtx.RUnlock()
//	res := make([]*User, 0, limit)
//	for _, u := range s.data {
//		if strings.Contains(u.Username, search) {
//			clone := u.User
//			res = append(res, &clone)
//		}
//		if len(res) == limit {
//			return res, nil
//		}
//	}
//	return res, nil
//}
//
//func (s *memStore) create(user *FullUserInfo) error {
//	if user == nil {
//		return nil
//	}
//
//	s.mtx.Lock()
//	defer s.mtx.Unlock()
//
//	if _, exists := s.data[user.Id.String()]; exists {
//		return errors.New("user already exists")
//	}
//
//	s.data[user.Id.String()] = cloneFullUserInfoStruct(user)
//
//	return nil
//}
//
//func (s *memStore) update(user *FullUserInfo) error {
//	if user == nil {
//		return nil
//	}
//
//	s.mtx.Lock()
//	defer s.mtx.Unlock()
//
//	if _, exists := s.data[user.Id.String()]; !exists {
//		return errors.New("user does not exists")
//	}
//
//	s.data[user.Id.String()] = cloneFullUserInfoStruct(user)
//
//	return nil
//}
//
//func (s *memStore) delete(id UUID) error {
//	s.mtx.Lock()
//	defer s.mtx.Unlock()
//
//	if _, exists := s.data[id.String()]; !exists {
//		return errors.New("user does not exists")
//	}
//
//	delete(s.data, id.String())
//
//	return nil
//}
//
////helper
//func cloneFullUserInfoStruct(u *FullUserInfo) *FullUserInfo {
//	if u.Id == nil {
//		return nil
//	}
//	clone := *u
//	if u.Id != nil {
//		clone.Id = append(make([]byte, 0, len(u.Id)), u.Id...)
//	}
//	if u.ActivationTime != nil {
//		activationTimeClone := *u.ActivationTime
//		clone.ActivationTime = &activationTimeClone
//	}
//	clone.ScryptSalt = append(make([]byte, 0, len(u.ScryptSalt)), u.ScryptSalt...)
//	clone.ScryptPwd = append(make([]byte, 0, len(u.ScryptPwd)), u.ScryptPwd...)
//	return &clone
//}
