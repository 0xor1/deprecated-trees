package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bytes"
	"fmt"
	"github.com/0xor1/isql"
	"strings"
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

func (s *sqlStore) getAccountRole(shard int, accountId, memberId Id) *AccountRole {
	return GetAccountRole(s.shards[shard], accountId, memberId)
}

func (s *sqlStore) setPublicProjectsEnabled(shard int, accountId, myId Id, publicProjectsEnabled bool) {
	_, err := s.shards[shard].Exec(`CALL setPublicProjectsEnabled(?, ?, ?)`, []byte(accountId), []byte(myId), publicProjectsEnabled)
	PanicIf(err)
}

func (s *sqlStore) getPublicProjectsEnabled(shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(s.shards[shard], accountId)
}

func (s *sqlStore) setMemberRole(shard int, accountId, myId, memberId Id, role AccountRole) {
	_, err := s.shards[shard].Exec(`CALL setAccountMemberRole(?, ?, ?, ?)`, []byte(accountId), []byte(myId), []byte(memberId), role)
	PanicIf(err)
}

func (s *sqlStore) getMember(shard int, accountId, memberId Id) *member {
	row := s.shards[shard].QueryRow(`SELECT id, isActive, role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := member{}
	PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
	return &res
}

/***
TODO need to determine which of these is most efficient on the db (this applys to project.GetMembers too):
1)
SELECT id, isActive, role
FROM accountMembers
WHERE account=:acc
AND isActive=true
AND (
        (
            name > (SELECT name FROM accountMembers WHERE account=:acc AND id=:id)
            AND role = (SELECT role FROM accountMembers WHERE account=:acc AND id=:id)
        )
        OR role > (SELECT role FROM accountMembers WHERE account=:acc AND id=:id)
)
ORDER BY role ASC, name ASC LIMIT :lim

2)
SELECT a1.id, a1.isActive, a1.role
FROM accountMembers a1, accountMembers a2
WHERE a1.account=:acc
AND a1.isActive=true
AND a2.account=:acc
AND a2.id=:id
AND (
        (
            a1.name>a2.name
            AND a1.role=a2.role
        )
        OR a1.role>a2.role
)
ORDER BY a1.role ASC, a1.name ASC LIMIT :lim
***/

func (s *sqlStore) getMembers(shard int, accountId Id, role *AccountRole, nameOrDisplayNameContains *string, after *Id, limit int) ([]*member, bool) {
	query := bytes.NewBufferString(`SELECT a1.id, a1.isActive, a1.role FROM accountMembers a1`)
	args := make([]interface{}, 0, 7)
	if after != nil {
		query.WriteString(`, accountMembers a2`)
	}
	query.WriteString(` WHERE a1.account=? AND a1.isActive=true`)
	args = append(args, []byte(accountId))
	if after != nil {
		query.WriteString(` AND a2.account=? AND a2.id=? AND ((a1.name>a2.name AND a1.role=a2.role) OR a1.role>a2.role)`)
		args = append(args, []byte(accountId), []byte(*after))
	}
	if role != nil {
		query.WriteString(` AND a1.role=?`)
		args = append(args, role)
	}
	if nameOrDisplayNameContains != nil {
		query.WriteString(` AND (a1.name LIKE ? OR a1.displayName LIKE ?)`)
		strVal := strings.Trim(*nameOrDisplayNameContains, " ")
		strVal = fmt.Sprintf("%%%s%%", strVal)
		args = append(args, strVal, strVal)
	}
	query.WriteString(` ORDER BY a1.role ASC, a1.name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func (s *sqlStore) getActivities(shard int, accountId Id, item *Id, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		InvalidArgumentsErr.Panic()
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, newValue FROM accountActivities WHERE account=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(accountId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
	}
	if occurredAfterUnixMillis != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfterUnixMillis)
	}
	if occurredBeforeUnixMillis != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBeforeUnixMillis)
	}
	if occurredAfterUnixMillis == nil && occurredBeforeUnixMillis == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*Activity, 0, limit)
	for rows.Next() {
		act := Activity{}
		PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.NewValue))
		res = append(res, &act)
	}
	return res
}

func (s *sqlStore) logActivity(shard int, accountId Id, member, item Id, itemType, action string, newValue string) {
	LogAccountActivity(s.shards[shard], accountId, member, item, itemType, action, nil, &newValue)
}
