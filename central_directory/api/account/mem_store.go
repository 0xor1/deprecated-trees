package account

import(
	. "bitbucket.org/robsix/task_center/misc"
	"sync"
	"strings"
)

func newMemStore() store {
	return &memStore{
		users: map[string]*fullUserInfo{},
		orgs: map[string]*org{},
		memberships: map[string]map[string]bool{},
		pwdInfos: map[string]*pwdInfo{},
		mtx: &sync.RWMutex{},
	}
}

type memStore struct{
	users map[string]*fullUserInfo
	orgs map[string]*org
	memberships map[string][]string //userId to orgIds
	pwdInfos map[string]*pwdInfo
	mtx *sync.RWMutex
}

func (s *memStore) copyFullUserInfo(user *fullUserInfo) *fullUserInfo {
	if user == nil {
		return nil
	}

	copy := *user
	if copy.Id != nil {
		copy.Id = Id(append(make([]byte, 0, 16), []byte(copy.Id)...))
	}
	if copy.NewRegion != nil{
		newRegionCopy := *copy.NewRegion
		copy.NewRegion = &newRegionCopy
	}
	if copy.NewEmail != nil{
		newEmailCopy := *copy.NewEmail
		copy.NewEmail = &newEmailCopy
	}
	if copy.ActivationCode != nil{
		activationCodeCopy := *copy.ActivationCode
		copy.ActivationCode = &activationCodeCopy
	}
	if copy.Activated != nil{
		activatedCopy := *copy.Activated
		copy.Activated = &activatedCopy
	}
	if copy.NewEmailConfirmationCode != nil{
		newEmailConfirmationCodeCopy := *copy.NewEmailConfirmationCode
		copy.NewEmailConfirmationCode = &newEmailConfirmationCodeCopy
	}
	if copy.ResetPwdCode != nil{
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
	if copy.NewRegion != nil{
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

func (s *memStore) getAccountByName(name string) (*account, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, user := range s.users {
		if user.Name == name {
			copy := s.copyFullUserInfo(user)
			val := account(copy.user)
			return &val, nil
		}
	}
	for _, org := range s.orgs {
		if org.Name == name {
			copy := s.copyOrg(org)
			val := account(*copy)
			return &val, nil
		}
	}
	return nil, nil
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
	delete(s.memberships, id.String())
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
	total := 0
	for _, user := range s.users {
		if strings.Contains(user.Name, search) {
			if total >= offset && len(res) < limit {
				copy := s.copyFullUserInfo(user)
				res = append(res, &copy.user)
			}
			total++
		}
	}
	return res, total, nil

}

func (s *memStore) createOrgAndMembership(user Id, org *org) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.orgs[org.Id.String()] = s.copyOrg(org)
	if s.memberships[user.String()] == nil {
		s.memberships[user.String()] = []string{org.Id.String()}
	} else {
		s.memberships[user.String()] = append(s.memberships[user.String()], org.Id.String())
	}
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
	for user, orgs := range s.memberships {
		for i := range orgs {
			orgs = orgs[:i+copy(orgs[i:], orgs[i+1:])]
		}
		s.memberships[user] = orgs
	}
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
	total := 0
	for _, org := range s.orgs {
		if strings.Contains(org.Name, search) {
			if total >= offset && len(res) < limit {
				res = append(res, s.copyOrg(org))
			}
			total++
		}
	}
	return res, total, nil
}

func (s *memStore) getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error) {
	usersOrgs := s.memberships[userId.String()]
	if usersOrgs == nil {
		return nil, 0, nil
	} else {

	}
}

func (s *memStore) membershipExists(user, org Id) (bool, error) {
	if s.memberships[user.String()] != nil && s.memberships[user.String()][org.String()] {
		return true, nil
	}
	return false, nil
}

func (s *memStore) createMembership(user, org Id) error {

}

func (s *memStore) deleteMembership(user, org Id) error {

}


