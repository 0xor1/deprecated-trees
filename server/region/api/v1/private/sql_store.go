package private

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bytes"
	"github.com/0xor1/isql"
	"math/rand"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		InvalidArgumentsErr.Panic()
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

func (s *sqlStore) createAccount(id Id, ownerId Id, ownerName string) int {
	shardId := rand.Intn(len(s.shards))
	_, err := s.shards[shardId].Exec(`CALL registerAccount(?, ?, ?);`, []byte(id), []byte(ownerId), ownerName)
	PanicIf(err)
	return shardId
}

func (s *sqlStore) deleteAccount(shard int, account Id) {
	_, err := s.shards[shard].Exec(`CALL deleteAccount(?);`, []byte(account))
	PanicIf(err)
}

func (s *sqlStore) getAllInactiveMemberIdsFromInputSet(shard int, accountId Id, members []Id) []Id {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId))
	query := bytes.NewBufferString(`SELECT id FROM accountMembers WHERE account=? AND isActive=false AND id IN (`)
	for i, mem := range members {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`?`)
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`);`)
	res := make([]Id, 0, len(members))
	rows, err := s.shards[shard].Query(query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	for rows.Next() {
		id := make([]byte, 0, 16)
		rows.Scan(&id)
		res = append(res, Id(id))
	}
	return res
}

func (s *sqlStore) getAccountRole(shard int, accountId, memberId Id) *AccountRole {
	row := s.shards[shard].QueryRow(`SELECT role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func (s *sqlStore) addMembers(shard int, accountId Id, members []*AddMemberPrivate) {
	queryArgs := make([]interface{}, 0, 3*len(members))
	query := bytes.NewBufferString(`INSERT INTO accountMembers (account, id, name, role) VALUES `)
	for i, mem := range members {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`(?, ?, ?, ?)`)
		queryArgs = append(queryArgs, []byte(accountId), []byte(mem.Id), mem.Name, mem.Role)
	}
	_, err := s.shards[shard].Exec(query.String(), queryArgs...)
	PanicIf(err)
}

func (s *sqlStore) updateMembersAndSetActive(shard int, accountId Id, members []*AddMemberPrivate) {
	for _, mem := range members {
		_, err := s.shards[shard].Exec(`UPDATE accountMembers SET name=?, role=?, isActive=true WHERE account=? AND id=?`, mem.Name, mem.Role, []byte(accountId), []byte(mem.Id))
		PanicIf(err)
	}
}

func (s *sqlStore) getTotalOwnerCount(shard int, accountId Id) int {
	count := 0
	IsSqlErrNoRowsAndPanicIf(s.shards[shard].QueryRow(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, []byte(accountId)).Scan(&count))
	return count
}

func (s *sqlStore) getOwnerCountInSet(shard int, accountId Id, members []Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId))
	query := bytes.NewBufferString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (`)
	for i, mem := range members {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`?`)
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`);`)
	count := 0
	IsSqlErrNoRowsAndPanicIf(s.shards[shard].QueryRow(query.String(), queryArgs...).Scan(&count))
	return count
}

func (s *sqlStore) setMembersInactive(shard int, accountId Id, members []Id) {
	accountIdBytes := []byte(accountId)
	for _, mem := range members {
		_, err := s.shards[shard].Exec(`CALL setAccountMemberInactive(?, ?)`, accountIdBytes, []byte(mem))
		PanicIf(err)
	}
}

func (s *sqlStore) renameMember(shard int, accountId Id, member Id, newName string) {
	_, err := s.shards[shard].Exec(`CALL renameMember(?, ?, ?)`, []byte(accountId), []byte(member), newName)
	PanicIf(err)
}

func (s *sqlStore) logActivity(shard int, accountId Id, member, item Id, itemType, action string) {
	LogAccountActivity(s.shards[shard], accountId, member, item, itemType, action, nil)
}

func (s *sqlStore) logAccountBatchAddOrRemoveMembersActivity(shard int, accountId, member Id, members []Id, action string) {
	LogAccountBatchAddOrRemoveMembersActivity(s.shards[shard], accountId, member, members, action)
}
