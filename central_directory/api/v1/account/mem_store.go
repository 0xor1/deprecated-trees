package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	q "github.com/ahmetalpbalkan/go-linq"
	"math"
	"strings"
	"sync"
)

type memStore struct {
	users           map[string]*fullUserInfo
	orgs            map[string]*org
	membershipsUtoO map[string]map[string]interface{} //userId to orgIds
	membershipsOtoU map[string]map[string]interface{} //orgId to userIds
	pwdInfos        map[string]*pwdInfo
	mtx             *sync.RWMutex
}

func (s *memStore) copyFullUserInfo(user *fullUserInfo) *fullUserInfo {
	if user == nil {
		return nil
	}

	copy := *user
	if copy.Id != nil {
		copy.Id = Id(append(make([]byte, 0, 16), []byte(copy.Id)...))
	}
	if copy.NewRegion != nil {
		newRegionCopy := *copy.NewRegion
		copy.NewRegion = &newRegionCopy
	}
	if copy.NewEmail != nil {
		newEmailCopy := *copy.NewEmail
		copy.NewEmail = &newEmailCopy
	}
	if copy.ActivationCode != nil {
		activationCodeCopy := *copy.ActivationCode
		copy.ActivationCode = &activationCodeCopy
	}
	if copy.Activated != nil {
		activatedCopy := *copy.Activated
		copy.Activated = &activatedCopy
	}
	if copy.NewEmailConfirmationCode != nil {
		newEmailConfirmationCodeCopy := *copy.NewEmailConfirmationCode
		copy.NewEmailConfirmationCode = &newEmailConfirmationCodeCopy
	}
	if copy.ResetPwdCode != nil {
		resetPwdCodeCopy := *copy.ResetPwdCode
		copy.ResetPwdCode = &resetPwdCodeCopy
	}

	return &copy
}

func (s *memStore) copyOrg(org *org) *org {
	if org == nil {
		return nil
	}

	copy := *org
	if copy.Id != nil {
		copy.Id = Id(append(make([]byte, 0, 16), []byte(copy.Id)...))
	}
	if copy.NewRegion != nil {
		newRegionCopy := *copy.NewRegion
		copy.NewRegion = &newRegionCopy
	}

	return &copy
}

func (s *memStore) copyPwdInfo(pwdInfo *pwdInfo) *pwdInfo {
	if pwdInfo == nil {
		return nil
	}

	copy := *pwdInfo
	if copy.Salt != nil {
		copy.Salt = append(make([]byte, 0, len(copy.Salt)), copy.Salt...)
	}
	if copy.Pwd != nil {
		copy.Pwd = append(make([]byte, 0, len(copy.Pwd)), copy.Pwd...)
	}

	return &copy
}

func (s *memStore) accountWithNameExists(name string) (bool, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, user := range s.users {
		if user.Name == name {
			return true, nil
		}
	}
	for _, org := range s.orgs {
		if org.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (s *memStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	idStr := user.Id.String()
	s.users[idStr] = s.copyFullUserInfo(user)
	s.pwdInfos[idStr] = s.copyPwdInfo(pwdInfo)
	return nil
}

func (s *memStore) getUserByName(name string) (*fullUserInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, user := range s.users {
		if user.Name == name {
			return s.copyFullUserInfo(user), nil
		}
	}
	return nil, nil
}

func (s *memStore) getUserByEmail(email string) (*fullUserInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, user := range s.users {
		if user.Email == email {
			return s.copyFullUserInfo(user), nil
		}
	}
	return nil, nil
}

func (s *memStore) getUserById(id Id) (*fullUserInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.copyFullUserInfo(s.users[id.String()]), nil
}

func (s *memStore) getPwdInfo(id Id) (*pwdInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.copyPwdInfo(s.pwdInfos[id.String()]), nil
}

func (s *memStore) updateUser(user *fullUserInfo) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.users[user.Id.String()] = s.copyFullUserInfo(user)
	return nil
}

func (s *memStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.pwdInfos[id.String()] = s.copyPwdInfo(pwdInfo)
	return nil
}

func (s *memStore) deleteUserAndAllAssociatedMemberships(id Id) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.users, id.String())
	delete(s.pwdInfos, id.String())
	for orgId := range s.membershipsUtoO[id.String()] {
		delete(s.membershipsOtoU[orgId], id.String())
	}
	delete(s.membershipsUtoO, id.String())
	return nil
}

func (s *memStore) getUsers(ids []Id) ([]*user, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*user, 0, len(ids))
	for _, id := range ids {
		if u := s.users[id.String()]; u != nil {
			copy := s.copyFullUserInfo(u)
			res = append(res, &copy.user)
		}
	}
	return res, nil
}

func (s *memStore) searchUsers(search string, offset, limit int) ([]*user, int, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	res := make([]*user, 0, limit)

	allMatches := q.From(s.users).Where(func(kv interface{}) bool {
		return strings.Contains(kv.(q.KeyValue).Value.(*fullUserInfo).Name, search)
	}).Select(func(kv interface{}) interface{} {
		copiedUser := s.copyFullUserInfo(kv.(q.KeyValue).Value.(*fullUserInfo)).user
		return &copiedUser
	})

	total := allMatches.Count()

	allMatches.OrderBy(func(u interface{}) interface{} {
		return u.(*user).Name
	}).Skip(offset).Take(limit).ToSlice(&res)

	return res, total, nil

}

func (s *memStore) createOrgAndMembership(user Id, org *org) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.orgs[org.Id.String()] = s.copyOrg(org)
	if s.membershipsUtoO[user.String()] == nil {
		s.membershipsUtoO[user.String()] = map[string]interface{}{org.Id.String(): nil}
	} else {
		s.membershipsUtoO[user.String()][org.Id.String()] = nil
	}
	s.membershipsOtoU[org.Id.String()] = map[string]interface{}{user.String(): nil}
	return nil
}

func (s *memStore) getOrgById(id Id) (*org, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.copyOrg(s.orgs[id.String()]), nil
}

func (s *memStore) getOrgByName(name string) (*org, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, org := range s.orgs {
		if org.Name == name {
			return s.copyOrg(org), nil
		}
	}
	return nil, nil
}

func (s *memStore) updateOrg(org *org) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.orgs[org.Id.String()] = s.copyOrg(org)
	return nil
}

func (s *memStore) deleteOrgAndAllAssociatedMemberships(id Id) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.orgs, id.String())
	for userId := range s.membershipsOtoU[id.String()] {
		delete(s.membershipsUtoO[userId], id.String())
	}
	delete(s.membershipsOtoU, id.String())
	return nil
}

func (s *memStore) getOrgs(ids []Id) ([]*org, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*org, 0, len(ids))
	for _, id := range ids {
		if o := s.orgs[id.String()]; o != nil {
			res = append(res, s.copyOrg(o))
		}
	}
	return res, nil
}

func (s *memStore) searchOrgs(search string, offset, limit int) ([]*org, int, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	res := make([]*org, 0, limit)

	allMatches := q.From(s.orgs).Where(func(kv interface{}) bool {
		return strings.Contains(kv.(q.KeyValue).Value.(*org).Name, search)
	}).Select(func(kv interface{}) interface{} {
		return kv.(q.KeyValue).Value.(*org)
	})

	total := allMatches.Count()

	allMatches.OrderBy(func(o interface{}) interface{} {
		return o.(*org).Name
	}).Skip(offset).Take(limit).ToSlice(&res)

	return res, total, nil
}

func (s *memStore) getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	usersOrgs := s.membershipsUtoO[userId.String()]
	if usersOrgs == nil {
		return nil, 0, nil
	} else {
		tmp := make([]*org, 0, len(usersOrgs))
		for orgId := range usersOrgs {
			tmp = append(tmp, s.orgs[orgId])
		}
		res := make([]*org, 0, int(math.Min(float64(limit), float64(len(usersOrgs)))))
		q.From(tmp).OrderBy(func(o interface{}) interface{} {
			return o.(*org).Name
		}).Skip(offset).Take(limit).ToSlice(&res)
		return res, len(usersOrgs), nil
	}
}

func (s *memStore) membershipExists(user, org Id) (bool, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	exists := false
	if s.membershipsUtoO[user.String()] != nil {
		_, exists = s.membershipsUtoO[user.String()][org.String()]
	}
	return exists, nil
}

func (s *memStore) createMembership(user, org Id) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.membershipsUtoO[user.String()] == nil {
		s.membershipsUtoO[user.String()] = map[string]interface{}{org.String(): nil}
	} else {
		s.membershipsUtoO[user.String()][org.String()] = nil
	}
	s.membershipsOtoU[org.String()][user.String()] = nil
	return nil
}

func (s *memStore) deleteMembership(user, org Id) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.membershipsUtoO[user.String()], org.String())
	delete(s.membershipsOtoU[org.String()], user.String())
	return nil
}
