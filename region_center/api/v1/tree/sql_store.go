package tree

//import (
//	. "bitbucket.org/robsix/task_center/misc"
//)
//
//func newSqlStore() store {
//	return &sqlStore{}
//}
//
//type sqlStore struct {
//}
//
//func (s *sqlStore) createTaskSet(*taskSet) (int, error) {
//
//}
//
//func (s *sqlStore) createMember(shard int, org Id, member *member) error {
//
//}
//
//func (s *sqlStore) deleteAccount(shard int, account Id) error {
//
//}
//
//func (s *sqlStore) getMember(shard int, org, member Id) (*member, error) {
//
//}
//
//func (s *sqlStore) addMembers(shard int, org Id, members []*NamedEntity) error {
//
//}
//
//func (s *sqlStore) setMembersActive(shard int, org Id, members []*NamedEntity) error {
//
//}
//
//func (s *sqlStore) getTotalOrgOwnerCount(shard int, org Id) (int, error) {
//
//}
//
//func (s *sqlStore) getOwnerCountInSet(shard int, org Id, members []Id) (int, error) {
//
//}
//
//func (s *sqlStore) setMembersInactive(shard int, org Id, members []Id) error {
//
//}
//
//func (s *sqlStore) memberIsOnlyOwner(shard int, org, member Id) (bool, error) {
//
//}
//
//func (s *sqlStore) setMemberDeleted(shard int, org Id, member Id) error {
//
//}
//
//func (s *sqlStore) renameMember(shard int, org Id, member Id, newName string) error {
//
//}
